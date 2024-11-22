# Development of an Industry-ready Kubernetes Operator

## Overview

In this article I would like to introduce the reader into the realms of cloud native technologies by showing how one can implement a fully custom operator using Go and Kubebuilder with a sprikle of React and lots of CLI tools for a fullstack experience. While I don't consider myself an expert in this field during my internship at Cisco I think acquired enough knowledge to show some useful things about this matter and hopefully dissolve some misconceptions about its difficulty.

In order to make this guide more similar the process one would do in reality when meaningful work is needed I would like to concentrate on things that go a bit further than simply achieving a "hello world" type of result. In this writing we will create an operator that uses Grype to acquire the current state of security of our Kubernetes Cluster.

The source code of the entire project is available on <https://github.com/kerezsiz42/scanner-operator2>.

## Containerization

Containerization is a lightweight virtualization method that packages applications and their dependencies into self-contained units called containers which enables efficient resource usage, rapid deployment, and easy scaling compared to VMs (virtual machines). The gain over traditional VMs is that the workloads are processes with some lightwight OS (operating system) APIs (application programming interfaces) for compatibility instead of full-blown OSs so this way they share a [common kernel](https://learn.microsoft.com/en-us/virtualization/windowscontainers/about/containers-vs-vm) of the host OS which achieves the benefitial properties.

### Kubernetes

It started out as a continuation to [Borg](https://medium.com/containermind/a-new-era-of-container-cluster-management-with-kubernetes-cd0b804e1409) which was Google's original internal Container Cluster Manager which at some point ran most of their systems. Kubernetes is nowadays the most used open source container orchestration tool which provides higher level concepts that Docker or other runtimes that conform to [OCI](https://opencontainers.org/) does not have, namely self-healing, manual and automatic horizontal scaling, load balancing and automated rollouts and rollbacks to name a few. Figuratively speaking it's the navigator for Docker containers hence its name means 'steersman' or 'pilot' (grc: κυβερνήτης).

- <https://www.paloaltonetworks.com/cyberpedia/containerization>
- <https://kubernetes.io/docs/concepts/overview/#going-back-in-time>

### Nodes and Clusters

Nodes are physical or virtual machines that provide computing resources for the Cluster they are part of in order to run applications and services. New Nodes can be added to a Cluster and existing can be removed. Hosted workloads can be arbitrarily moved from one Node to other Nodes when performing maintenance.
At a minimum, a cluster contains a control plane and one or more compute machines, or nodes. The control plane is responsible for maintaining the desired state of the cluster, such as which applications are running and which container images they use. Nodes actually run the applications and workloads. See: <https://www.redhat.com/en/topics/containers/what-is-a-kubernetes-cluster>

### Namespace

Namespaces are Cluster-wide unique resources which are used to create a virtual separation between components. The name, namespace and kind triad uniquely identifies every resource within a Cluster. We will take advantage of this later when I discuss about the operator.

### Pods, ResplicaSets and Deployments

Pods are arguably the most important resource within Kubernetes, as they incorporate one or more containers which can potentially share storage and network resources. They are considered the smallest deployable unit of computing since Kubernetes does not manage containers alone. Other important resorces are ReplicaSets which make sure that the requested number identical Pods are running at a given time and Deployments which incorporate ReplicaSets and provide mechanisms for rolling updates and rollbacks to minimize the downtime of an application.

### Services and Ingress

Each pod has a unique IP address, so if one group of Pods wants to communicate with another, all the other Pods would have to know how to reach them, but the number of Pods can change dynamically due to their ephemeral nature. This problem is solved by the Service resource, using which we can provide a facade or abstraction over a multiple of Pods that are selected by tags.

An Ingress is a resource used to manage external access to services within a Kubernetes cluster, typically for HTTP and HTTPS traffic. It allows you to define rules for routing client requests to the appropriate services, based on hostnames, paths, or other factors. Unlike other methods of exposing services, such as NodePort or LoadBalancer, Ingress consolidates routing rules into a single point of control, which simplifies the management of traffic to multiple services. An Ingress controller, which operates as a reverse proxy, is necessary to implement these rules and manage load balancing, SSL termination, and traffic routing within the cluster.

## Why do we need Operators?

Controllers in Kubernetes are automations that have access to the Kubernetes API and other resources - often outside the cluster - in order to observe the state of the cluster and act on the changes in accordance with the logic they were implemented with. Operators are basically controllers which define a CRD (custom resource defintion), so they are effectively a way to extend the functionality of Kubernetes.

Read more:

- <https://kubernetes.io/docs/concepts/architecture/controller/>
- <https://konghq.com/blog/learning-center/kubernetes-controllers-vs-operators>

## Tooling Setup

### Go and Kubectl

First we have to make sure we have the most recent stable version of go and kubectl. Kubectl is basically a command line interface tool, with the help of which we can run commands in a given Cluster. In this case, we essentially communicate with the Kubernetes API server component via its the REST protocol.
Each operating system has its own way of installing and managing these packages, but if you not need the newest version because of a certain new feature then it's more convenient to just rely upon the package provided by your distribution instead of what a dedicated version manager like [gvm](https://github.com/moovweb/gvm) can provide. As long as the positives don't outweigh the amount of extra work we have to put into managing things, we should go with the default option for simplicity.

```sh
$ go version
go version go1.23.0 linux/amd64
```

```sh
$ kubectl version
Client Version: v1.30.4
Kustomize Version: v5.0.4-0.20230601165947-6ce0bf390ce3
Server Version: v1.31.0
```

Furthermore, we will need helm which I will introduce later.

```sh
$ helm version
version.BuildInfo{Version:"v3.16.1", GitCommit:"5a5449dc42be07001fd5771d56429132984ab3ab", GitTreeState:"clean", GoVersion:"go1.22.7"}
```

### Kubebuilder

We wil use Kubebuilder which is a framework for building Kubernetes APIs using custom resource definitions (CRDs). It streamlines the process of developing Kubernetes controllers by providing a set of tools, libraries, and code generation features. Built on top of the Kubernetes controller-runtime library, it allows developers to create robust, scalable controllers and operators with minimal boilerplate code. It supports best practices like testing, scaffolding, and project layout, making it easier to manage and extend Kubernetes-native applications.
We can download the latest release of Kubebuilder CLI, make it executable and move it into `/usr/local/bin` where user installed binaries are usually stored.

```sh
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/
```

After that we should be able to use Kubebuilder.

```sh
$ kubebuilder version
Version: main.version{KubeBuilderVersion:"4.1.1", KubernetesVendor:"1.30.0", GitCommit:"e65415f10a6f5708604deca089eee6b165174e5e", BuildDate:"2024-07-23T07:11:14Z", GoOs:"linux", GoArch:"amd64"
```

Binaries installed by the `go install` command are placed into `$(go env GOPATH)/bin` which is usually equal to `~/go/bin`.
We can make these commands callable by putting this directory onto our `$PATH` if needed.

```sh
printf '\nexport PATH="$(go env GOPATH)/bin:$PATH"\n' >> ~/.zshrc
exec zsh # Reinitializing our shell in order to make this change effectful
```

After that we can initialize a new project with a domain name of our choice and a github repository URL which will be the name of our Go moudle that we will be working on.

```sh
kubebuilder init --domain zoltankerezsi.xyz --repo github.com/kerezsiz42/scanner-operator2
```

Then we create an API with a group, version and kind name.

```sh
kubebuilder create api --group scanner --version v1 --kind Scanner
```

After these steps we will have a scaffold generated in the `internal/controller` folder for us. The two most important methods are the `Reconciler` and the `SetupWithManager` which we will take a more in depth look at later.

### Kind

Also, in order to setup a cluster we can use a tool called [kind](https://kind.sigs.k8s.io/) which makes it easy to create and delete Kubernetes clusters and nodes for our development. When using kind the nodes are running as docker containers instead of VMs in the cloud.

```sh
go install sigs.k8s.io/kind@v0.24.0
kind create cluster
```

We can verify if `kubectl` is configured correctly for this cluster.

```sh
$ kubectl cluster-info --context kind-kind
Kubernetes control plane is running at https://127.0.0.1:37675
CoreDNS is running at https://127.0.0.1:37675/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

## Setting up the Operator and its HTTP API

### Implementing new Logic

We can modify the `ScannerReconciler` structure and its methods to implement our initial testing logic, explore the workings of `kubebuilder` and to get familiar with the development workflow. Firstly, as we will need a separate HTTP API outside the one used for metrics, we can add a `http.Server` field to the `ScannerReconciler` struct which is zero-initialized (which in this case means that it will be `nil` in the begining).

```go
// internal/controller/scanner_controller.go

type ScannerReconciler struct {
  client.Client
  Scheme *runtime.Scheme
  Server *http.Server
}
```

The Kubebuilder team kindly marked as the place where we should put our code, so we will start the HTTP server there. Here we define a single handler function that returns the string `"Hello, world!"` and start the server in a `goroutine`, since `http.Server.ListenAndServe()` is a blocking call. We use `os.Exit(1)` together with `log.Error()` instead of `panic()` the same way as in `cmd/main.go` for failures that make further continuation of the process impossible. We intentionally let the process die, since it will be restarted by Deployment anyway. Also, the convention is to start the error messages with lowercase letters. After the first reconciliation is done the HTTP server should be running on port `8000` and the logs of the controller-manager pod should contain the `"successfully reconciled"` message.

```go
// internal/controller/scanner_controller.go

func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
  reconcilerLog := log.FromContext(ctx)

  if r.Server == nil {
    mux := http.NewServeMux()
    mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      if _, err := fmt.Fprintf(w, "Hello, world!\n"); err != nil {
        reconcilerLog.Error(err, "error while handling request")
        os.Exit(1)
      }
    }))

    r.Server = &http.Server{Addr: ":8000", Handler: mux}

    go func() {
      if err := r.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        reconcilerLog.Error(err, "unable to start HTTP server")
        os.Exit(1)
      }
    }()
  }

  reconcilerLog.Info("successfully reconciled")

  return ctrl.Result{}, nil
}
```

### Kustomize

Kustomize is a tool for customizing Kubernetes configurations. It allows us to generate resources from other resources and setting cross-cutting fields for resources along with composing and customizing collections of resources as documented on the official [Kubernetes docs website](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#overview-of-kustomize). It is used here to make a more concise version of the Kubernetes resource files by minimizing the amount of copied parts thereby simplifing maintenance. In this project the `config` folder is full of such resource definitions.

In the `config/default/kustomization.yaml` file we can set the namespace where the controller resources will be placed and a name prefix for these resources. Furthermore the `resources` section contains references to other kustomization files that will need to processed such as the RBAC (role based access control) rules and the CRD itself. Even though we will only have a single instance of the controller-manager, it is still a good practice to create a Service for it. To achieve that we will create a new file named `config/default/api_service.yaml` and add it to the kustomization resources. The final service looks like the following:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: controller-manager
    app.kubernetes.io/name: scanner-operator2
    app.kubernetes.io/managed-by: kustomize
  name: controller-manager-api-service
  namespace: system
spec:
  ports:
    - name: http
      port: 8000
      protocol: TCP
      targetPort: 8000
  selector:
    control-plane: controller-manager
```

So thanks to kustomize name will be prefixed with `scanner-` and the namespace will be replaced with `scanner-system` once the resource is applied.

### Development Workflow

We can change the default `imagePullPolicy` from `Always` to `IfNotPresent` implicitly by changing the image tag to `dev` instead of `latest` in our Makefile. Doing this will prevent Kubernetes to always try to pull the image when we run `make deploy` later. We will also set a proper Github Container Registry address, so that later we can push the built image there to be used by the packaged version of our software.

```Makefile
IMG ?= ghcr.io/kerezsiz42/scanner-operator2:dev
```

```Makefile
.PHONY: kind-load
kind-load:
  kind load docker-image ${IMG}
```

```sh
make docker-build # Build the docker image
make docker-push # Optionally push the image to the registry. This step is only needed if we want to test the helm deployment.
make kind-load # Load the new docker image into kind cluster
make deploy # Deploy or redeploy all resources that are needed for the newest version
```

```sh
kubectl api-resources --verbs=list -o name | grep scanner
```

After this there should be pod a named scanner-controller-manager with `Running` status, along with the new service we just defined, the metric service, deployment and replicaset.

```sh
$ kubectl get all -n scanner-system
NAME                                              READY   STATUS    RESTARTS   AGE
pod/scanner-controller-manager-777b874846-c2tpw   1/1     Running   0          25m

NAME                                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/scanner-controller-manager-api-service       ClusterIP   10.96.21.245    <none>        8000/TCP   25m
service/scanner-controller-manager-metrics-service   ClusterIP   10.96.174.233   <none>        8443/TCP   25m

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/scanner-controller-manager   1/1     1            1           25m

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/scanner-controller-manager-777b874846   1         1         1       25m
```

We can access the logs outputted by the controller-manager using the `kubectl logs` command. Once we made sure the controller-manager pod is running properly, we can apply our new scanner resource described in the file at `config/samples/scanner_v1_scanner.yaml`. Then to manually test the API we will use the `kubectl port-forward` command to make the previously defined service accessible on our host machine.

```sh
kubectl apply -f config/samples/scanner_v1_scanner.yaml
kubectl port-forward service/scanner-controller-manager-api-service -n scanner-system 8000:8000
```

Using the `curl` command we can verify the result from our 'backend'.

```sh
$ curl localhost:8000
Hello, world!
```

### Helm and Helmify

Helm is today the industry standard package manager for Kubernetes, so we will use this to create our packaged operator that can later be downloaded deployed and undeployed in a Kubernetes namespace similarly to [Istio](https://istio.io/latest/docs/ambient/install/helm/) without the need of the source code and running `make install` and `make deploy` manually on every change.

[Helmify](https://github.com/arttor/helmify) is a tool that creates Helm charts from Kubernetes manifests (the yaml files). When running `make helm` it generates a helm chart in the `chart` directory of our repository. Our work here consists of copying the right Makefile commands from the documentation and running them appropriately when we create a new version of our software.

```Makefile
.PHONY: helm
helm: manifests kustomize helmify
  $(KUSTOMIZE) build config/default | $(HELMIFY)
```

After this we should be able to try out the deployment using helm. In order to start with a clean slate, we can recreate the cluster and create manually the new scanner-system namespace which will be used by the chart, and install it there. Of course kind does not have access to the docker image so we will have to load it again.

```sh
make helm # Generate the up-to-date helm chart within the ./char folder
kind delete cluster && kind create cluster
# If the image is not present, then pull it: "docker pull ghcr.io/kerezsiz42/scanner-operator2:dev"
make kind-load # Load image into the kind cluster
kubectl create namespace scanner-system
helm install scanner ./chart -n scanner-system # Or "helm install scanner --repo https://github.com/kerezsiz42/scanner-operator2/tree/main/chart -n scanner-system"
```

We can verify that it is working the same way as before and call the base endpoint that replies with the `"Hello, world!"` message exactly how we programmed.

```sh
kubectl apply -f config/samples/scanner_v1_scanner.yaml
kubectl port-forward service/scanner-chart-controller-manager-api-service -n scanner-system 8000:8000
$ curl localhost:8000
Hello, world!
```

To remove the deployment we run the `helm uninstall` command, and optionally remove the namespace too. By using helm we can potentially make use of its many benefits, which are dependency management, release management and parameterization. All these features are necessary to consider our project ready for industrial use.

```sh
helm uninstall scanner -n scanner-system
kubectl delete namespace scanner-system
```

## Setting up the Frontend Development Environment

The UI that we assemble here is considered to be the test or proof that the operator does what it has to and with a reasonable performance. Discussing this here might be considered a deviation from the original goal, but for the sake of testing it might still be worth implementing a proper UI.

For simplicity's sake, I choose to develop this UI using React, since the tooling around it is very mature. Also, we will be using Node instead of the newer more modern javascript runtimes like Bun or Deno. These are functionally mostly compatible with Node but there could still be some rough edges or surprizing hardships when you are trying to achieve an exact result, or at least this is true at the time of this writing.

Node can also be installed with the preferred version manager like [nvm](https://github.com/nvm-sh/nvm) or [fnm](https://github.com/Schniz/fnm). I will be using the latest release at the time of writing this document.

```sh
$ node --version
v22.6.0
```

First thing we should do is to initialize the project within the a frontend directory and install the necessary dependencies using npm.

```sh
mkdir frontend
cd frontend
npm init -y
npm install esbuild react react-dom @types/react-dom tailwindcss
```

- [esbuild](https://esbuild.github.io/) is a fast bundler for Javascript and Typescript. A bundler is a tool that takes multiple source code files and combines them into one or more depending on the configuration. Before the `import` directive was available, using a bundler was our only choice to ship code with multiple dependencies if we did not want to use global scoped objects as with [JQuery](https://jquery.com/). If you read the whole thesis, I owe you a drink. Now that [module syntax](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules) is standardized, it is still useful to bundle our frontend code to minimize the number of network requests the browser has to make in order to gather all of the source files and run our code. The other reason is of course that we use Typescript which is directly not runnable by the browser, so a tranformation step is necessary.
- `react`, `react-dom` and `@types/react-dom` are the packages we need to use to have all the necessary components of React for the web when we use Typescript.
- `tailwindcss` is a [utility first](https://tailwindcss.com/docs/utility-first) CSS compiler that has a purpose similar to a Javascript bundler. Looks for files specified by the pattern in tailwind.config.js and searches for existing Tailwind class names specified in those, in order to include them in the final `output.css`.

After that, we create the entrypoint of our single page application, the `index.html` file. The `<script />` tag together with the defer keyword is used to load the compiled frontend code witch will take over the `<div />` element which has the id of `app`, once the complete html file has been received and attached to the DOM. Cascading style sheets are loaded too using the `<link />` tag within `<head />`.

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link href="output.css" rel="stylesheet" />
    <script defer src="bundle.js"></script>
  </head>
  <body>
    <div id="app"></div>
  </body>
</html>
```

The initial version of our client-side code to test our frontend setup looks like the following.

```tsx
import * as React from "react";
import ReactDOM from "react-dom/client";

const element = document.getElementById("app")!;

const root = ReactDOM.createRoot(element);

function App() {
  return <h1 className="text-3xl font-bold underline">Hello, world!</h1>;
}

root.render(<App />);
```

We also define some script in our `package.json` file to document the steps it takes to build the final javascript and CSS files which can later be copied and served using a HTTP server. Here we are using the "production" `NODE_ENV` enviroment variable which instructs the bundler to omit program code that would enable us to attach certain debugging tools like the React Developer Tools that makes browsers aware of reacts internal state and behavior. This value can be changed back anytime for debugging purposes.

```json
{
  "scripts": {
    "build-css": "./node_modules/.bin/tailwindcss -i src/input.css -o output.css",
    "build-js": "./node_modules/.bin/esbuild src/index.tsx --define:process.env.NODE_ENV=\\\"production\\\" --bundle --outfile=bundle.js",
    "build": "npm run build-css && npm run build-js"
  }
}
```

The initial UI looks like the following in it's rendered form.

![Initial rendered UI](initial-ui.png)

With all this done we can confirm that with this setup we can develop a modern UI with a fast and easy to understand iteration loop since all components are configured correctly to get the resulting Javascript, CSS and HTML files.

## Exploration of the Idea

The goal of this operator is to give constant feedback about the security status of our Kubernetes Cluster. It collects and scans container images using an external tool that can run in a Pod and consequently as a Job. In our test, we will install the custom resource in the default namespace.í

The controller-manager will be subscribed to certain events and therefore notified by the Kubernetes API each time a new Pod is started and it will look into the connected database in order to decide if it should start a new scan Job or not. Successful Jobs return their results by calling an endpoint on the REST API that is provided by the operator. The service that connects to the operator deployment can be accessed from the outside either through an ingress or by using port-forwarding.

![Architecture](architecture.png)

We will make use of the publisher-subscriber architectural pattern which is an often used communication method between components of distributed systems. The advantage of this approach is the loose coupling and scalability it provides and that subscribers will get a faster - almost real-time - notification when an event occurs compared to polling wich would involve the client calling the API repeatedly thereby cause unnecessary traffic. When a client loads the frontend code, it automatically subscribes to scan events by using Websocket. Once such event occurs the client can load the result from the same REST API and display them in a list on the user interface.

![Sequence Diagram](sequence-diagram.png)

### Grype and CVEs

We will be using [Grype](https://github.com/anchore/grype) which is a vulnerability scanner for container images and filesystems. It will do essence of the work. A scanning process results in some collection of [CVEs](https://www.cve.org/), which draw our attention to weaknesses in computational logic found in software and hardware components that, when exploited results in a negative impact to confidentiality, integrity, or availability of our product.

More info:

- <https://nvd.nist.gov/vuln>

It supports multiple type of outputs, but from them [OWASP CycloneDX](https://cyclonedx.org/specification/overview/) [SBOM](https://www.cisa.gov/sbom) (software bill of materials) - an object model which shows a nested inventory or a list of ingredients that make up software components - contains probably the most information, so we will use that. Furtunately there is a Go library available to us that supports this format, so we can add it to our project dependencies:

```sh
go get github.com/CycloneDX/cyclonedx-go
```

```sh
$ grype version
Application:         grype
Version:             0.82.0
BuildDate:           2024-10-07T21:36:44Z
GitCommit:           6b09bb857564cd3c59c0cc1b6ea997c5ee198b6d
GitDescription:      v0.82.0
Platform:            linux/amd64
GoVersion:           go1.23.2
Compiler:            gc
Syft Version:        v1.14.0
Supported DB Schema: 5
```

When trying Grype multiple times using a kubernetes Pod a potential pitfall becomes obvious: grype has to download its vulnerability database in each pod before performing the scan, which becomes an average 30 second long delay for each scan instead of a single initial delay. This is very wasteful. Trying the same with the local CLI does not produce the same issue, since this database is cached in the filesystem, although running two at the same time scans before a database is present results in the failure of both.

```sh
kubectl run -it grype --image=anchore/grype -- python
kubectl delete pod grype
kubectl run -it grype --image=anchore/grype -- python # Downloads vulnerability db again
```

Upon further inspection we can see that grype does provide a way to manage its database through subcommands, so this might be an additional responsibility that our operator could be in charge of, but then it might be a better solution to avoid running jobs concurrently. Doing it that way makes throttling unnecessary, because not putting too much processing load on the cluster is also an important aspect of our goal.

<https://stackoverflow.com/questions/62694361/how-to-reference-a-local-volume-in-kind-kubernetes-in-docker>

### OpenAPI and REST

REST stands for respresentation state transfer and is a set of architectural constraints that makes the job of designers, developers and users of an application programming interface easier by providing a few loose rules to follow. RESTful systems are stateless, cacheable, have a layered structure and when paired with client applications their inner workings are entirely decoupled.

In order to comply with today's standards we implement the Backend API using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/) which is a tool that allows us to generate boilerplate go code out of an OpenAPI 3.0 definition (here `oapi_definition.yaml`). This way we can spend more time developing our business logic.

There also exists a client-side code generator for Typescript ([openapi-ts](https://openapi-ts.dev/introduction)) along with a matching HTTP client, but since the API will be very simple, we will only install the code generator to automatically generate the reponse objects for us that we can later arbitrarily place anywhere.

```sh
npm i openapi-typescript
```

After installing the new package we can set up an npm script so that we do not have to remember the command every time:

```json
// package.json
{
  "scripts": {
    "gen": "./node_modules/.bin/openapi-typescript ../internal/oapi/oapi_definition.yaml -o ./src/oapi.gen.d.ts"
  }
}
```

## Synchronizing Front- and Backend Functionality

In this section we will install the remaining dependencies and establish communication between the client and server side, setup and manually test the database connection.

Contrary to my initial plans instead of Webhooks we will use Websockets. The reason for this is the performance gain that a continously open TCP connection can provide over sending new and new HTTP requests for each scanned image to the subscribers. Websocket is bidirectional by nature, but we will use it in a unidirectional way to notify clients about updates.

The following component is a facade over Websocket client, which implements automatic reconnection and extends [EventTarget](https://developer.mozilla.org/en-US/docs/Web/API/EventTarget) to provide access to its "message" events. Using the `addEventListener` method one can register a callback that can process the incoming `CustomEvent` object as it arrives. The connection can be closed using the web standard [AbortSignal API](https://developer.mozilla.org/en-US/docs/Web/API/AbortSignal). It works with the assumption that the publisher will always send JSON strings so it tries to parse the incoming data accoringly.

![Subscriber Class](subscribe-class.png)

```js
//@ts-check

/**
 * @typedef {{signal?: AbortSignal}} SubscriberOptions
 */

export class Subscriber extends EventTarget {
  /** @type {WebSocket | undefined} */
  #ws = undefined;
  /** @type {number | undefined} */
  #timeoutId = undefined;
  #isConnected = false;
  #previousIsConnected = false;
  #shouldBeOpen = true;

  /**
   * @param {string} url
   * @param {SubscriberOptions} [options]
   */
  constructor(url, options) {
    super();
    this.#ws = this.#connect(url);

    options?.signal?.addEventListener("abort", () => {
      this.#shouldBeOpen = false;
      this.#ws?.close();
    });
  }

  /**
   * @param {boolean} newState
   */
  #setState(newState) {
    this.#previousIsConnected = this.#isConnected;
    this.#isConnected = newState;

    if (this.#previousIsConnected !== this.#isConnected) {
      const ce = new CustomEvent("connection", { detail: this.#isConnected });
      this.dispatchEvent(ce);
    }
  }

  /**
   * @param {string} url
   * @returns {WebSocket}
   */
  #connect(url) {
    clearTimeout(this.#timeoutId);
    const ws = new WebSocket(url);

    ws.onopen = () => {
      this.#setState(true);
    };

    /**
     * @param {MessageEvent<string>} ev
     */
    ws.onmessage = (ev) => {
      const detail = JSON.parse(ev.data);
      const ce = new CustomEvent("message", { detail });
      this.dispatchEvent(ce);
    };

    ws.onclose = (_ev) => {
      this.#setState(false);
      if (this.#shouldBeOpen) {
        this.#timeoutId = setTimeout(() => this.#connect(url), 5000);
      }
    };

    return (this.#ws = ws);
  }
}
```

This class is documented using JSDoc type annotations which can be checked using the same Typescript compiler like normal `.ts` files so the two solutions are basically equivalent, except it is a valid `.js` file and the type annotations are comments. As a test, we can define an anonymous function in a `useEffect` hook. We create an AbortController that can be used to emit an abort signal when the component gets unmounted from the DOM and register a function on the "message" event so it prints the content to the console for now.

```tsx
// frontend/src/index.tsx
useEffect(() => {
  const ac = new AbortController();
  const s = new Subscriber("/subscribe", { signal: ac.signal });
  s.addEventListener("message", (e: CustomEventInit) => {
    console.log(e.detail);
  });

  return () => ac.abort();
}, []);
```

On the client side fortunately all modern browsers implement this feature, but since the Go standard library does not implement a Websocket server we need to introduce a new dependency: Gorilla Websocket

```sh
go get github.com/gorilla/websocket
```

We can create a new package and its folder called `oapi` within the `internal` folder next to the controller, which will serve as a place to store everything that concerns itself with OpenAPI and the code generation itself.

```sh
$ tree internal/oapi
internal/oapi
├── config.yaml
├── generate.go
├── oapi_definition.yaml
├── oapi.gen.go
└── oapi.go

1 directory, 5 files
```

Here we are using the recommended 'tools go' approach, which means that instead of installing the `oapi-codegen` code generator as a binary external to our program we introduce it into the version control system of `go.mod` in order to manage it as a dependency alongside our core application.

```go
// internal/oapi/oapi.go

//go:build oapi
// +build oapi

package oapi

import (
  _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
```

This effectively enables us to call this CLI program through `go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` with the `--config` parameter and the path to our definition. In the `generate.go` we make use of the `go:generate` directive and Go's metaprogramming capabilities, which means that we can write program code using program code. This process is ran by the programmer before compilation or build time in other words during development.

```go
// internal/oapi/generate.go
package oapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml oapi_definition.yaml
```

Nonetheless, we still need to include the runtime library of `oapi-codegen` which contains code that our generated code can make use of.

```sh
go get github.com/oapi-codegen/runtime
```

We can configure it to only generate models out of OpenAPI schemas and the `ServerInterface` with the proper methods that correspond to the paths section in our definition. Running the `make gen` command creates the `oapi.gen.go` file.

```yaml
# internal/oapi/config.yaml
package: oapi
generate:
  std-http-server: true
  models: true
output: oapi.gen.go
```

First we define a new API endpoint in our OpenAPI 3.0 specification. The info section provides metadata, including the API version (1.0.0) and the title (Scanner Operator API). Under paths, the root path (/) is defined with a GET operation. The response for a successful GET request (HTTP status 200) is specified to return HTML content, with the content type as text/html, and the response schema being a simple string. We will create such definitions for all of our endpoints including the WebSocket too. A great resource that can help us when specifying such definitions is the official [Swagger](https://swagger.io/docs/specification/v3_0/describing-parameters/) documentation website.

```yaml
# internal/oapi/oapi_definition.yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Scanner Operator API
paths:
  /:
    get:
      responses:
        "200":
          description: Return the HTML part of the frontend
          content:
            text/html:
              schema:
                type: string
# ...
```

Once the new `oapi.gen.go` file is created we can make use of these `structs` and methods in our implementation class. Go's type safety ensures that in our implementation all methods are defined in accordance with the generated `ServerInterface`. I case our implementation is incorrect the compiler wil show the type mismatch at `oapi.HandlerFromMux()` where the server implementation is used.

```go
// internal/server/server.go
// ...

type Server struct {
  db       *gorm.DB
  upgrader *websocket.Upgrader
}

func NewServer(db *gorm.DB) *Server {
  return &Server{
    db:       db,
    upgrader: &websocket.Upgrader{},
  }
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  w.WriteHeader(http.StatusOK)
  _, _ = w.Write(frontend.IndexHtml)
}

// ...
```

Go modules are created by a go.mod file which contain the module path, which - as generated by kubebuilder - the github repo path to our project. This way other modules can reference and download ours through github in order to use it. Modules contain packages (the folders) and packages contain the source files. Source files within a package have shared scope, so one file can access constants, variables, functions and structs without additional importing.

Static files can be built into the binary as byte slices using the embed package, which is one of Go's unique features. We can place the `embed.go` file into the `frontend` folder, and reference these variables under the frontend package name in `internal/server` where we will build them into HTTP response handlers.

```go
// frontend/embed.go
package frontend

import _ "embed"

//go:embed bundle.js
var BundleJs []byte

//go:embed index.html
var IndexHtml []byte

//go:embed output.css
var OutputCss []byte
```

Of course all source files should be added to the final build so we will have to make sure these a are copied or added in the `Dockerfile` properly.

```Dockerfile
# ...
COPY frontend/bundle.js frontend/bundle.js
COPY frontend/index.html frontend/index.html
COPY frontend/output.css frontend/output.css
COPY frontend/embed.go frontend/embed.go
# ...
```

To the the WebSocket connection we can define `GetSubscribe`, which handles WebSocket connections for a server. It upgrades an HTTP request to a WebSocket connection using `s.upgrader.Upgrade`, and if an error occurs, it logs the error and exits. Once the connection is successfully established, the function enters a loop where it sends the message "hello" to the client every 3 seconds using `c.WriteMessage`. If writing to the WebSocket connection fails, it logs the error and breaks the loop. The connection is closed once the function completes, thanks to the `defer c.Close()` statement.

```go
// internal/server/server.go
// ...

func (s *Server) GetSubscribe(w http.ResponseWriter, r *http.Request) {
  c, err := s.upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  for {
    data := []byte("\"hello\"")
    if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
      log.Print("write:", err)
      break
    }

    time.Sleep(3 * time.Second)
  }
}
```

### Setting Up the Database Connection

In order to keep the project simple we are not going to implement complex database migration capabilites, so ![Gorm](https://gorm.io/)'s `AutoMigrate` is more than enough. Essentially we check on every startup if the database is clean and follow through with the migration. Otherwise the service return an error that the end-user would eventually see.

```go
// internal/database/dataabse.go
func GetDatabase() (*gorm.DB, error) {
  databaseType := os.Getenv("DATABASE_TYPE")
  dsn := os.Getenv("DSN")

  dialector, err := GetDialector(databaseType, dsn)
  if err != nil {
    return nil, fmt.Errorf("failed to get dialector: %w", err)
  }

  db, err := gorm.Open(dialector, &gorm.Config{})
  if err != nil {
    return nil, fmt.Errorf("failed to connect to database: %w", err)
  }

  if err := db.AutoMigrate(&Hello{}); err != nil {
    return nil, fmt.Errorf("failed to automigrate database: %w", err)
  }

  return db, nil
}
```

Here the `GetDialector` is just a switch construct that based on string comparison returns the correct dialector instance - the component that defines what type of sql database Gorm will communicate with. Another thing to note, is that the `GetDatabase` also accesses the `DATABASE_TYPE` and the `DSN` (data source name) environment variables internally.

We also have to make sure the Gorm library and the drivers that we plan to use with are all installed:

```sh
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/postgres
```

### Testing the Current Setup

After these modifications the `Reconcile` method checks if the database connection (r.Db) is nil, it tries to establish one using database.GetDatabase() and exits if it fails. Similarly, if the HTTP server (r.Server) is not running, it creates a new server instance and starts it asynchronously. Errors during server startup are logged, and the application exits if the server cannot be started. If both components are successfully initialized, the method logs a successful reconciliation and returns without errors.

```go
// internal/controller/scanner_controller.go
func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
  reconcilerLog := log.FromContext(ctx)

  if r.Db == nil {
    reconcilerLog.Info("connecting to database")
    db, err := database.GetDatabase()
    if err != nil {
      reconcilerLog.Error(err, "unable to connect to database")
      os.Exit(1)
    }

    r.Db = db
  }

  if r.Server == nil {
    si := server.NewServer(r.Db)
    m := http.NewServeMux()

    r.Server = &http.Server{
      Handler: oapi.HandlerFromMux(si, m),
      Addr:    ":8000",
    }

    go func() {
      reconcilerLog.Info("starting HTTP server")
      if err := r.Server.ListenAndServe(); err != http.ErrServerClosed {
        reconcilerLog.Error(err, "unable to start HTTP server")
        os.Exit(1)
      }
    }()
  }

  reconcilerLog.Info("successfully reconciled")

  return ctrl.Result{}, nil
}
```

To create a test database that our operator can connect to we can define the following deployment and service:

- The deployment deploys a single replica of a PostgreSQL container (postgres:17-alpine) using an image from Docker Hub. The container exposes port 5432, the default PostgreSQL port. It sets up environment variables for the database user (`POSTGRES_USER`), password (`POSTGRES_PASSWORD`), and database name (`POSTGRES_DB`). It uses an emptyDir volume mounted at /cache inside the container, which will persist data for the lifetime of the pod.

- The service exposes the PostgreSQL container to other services in the cluster via a Kubernetes Service named postgres-service. The service forwards traffic on port `5432` (TCP) to the PostgreSQL pod, enabling access to the database.

```yaml
# postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: postgres
  template:
    metadata:
      labels:
        app.kubernetes.io/name: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:17-alpine
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_PASSWORD
              value: password
            - name: POSTGRES_USER
              value: user
            - name: POSTGRES_DB
              value: db
          volumeMounts:
            - mountPath: /cache
              name: postgres-volume
      volumes:
        - name: postgres-volume
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
spec:
  selector:
    app.kubernetes.io/name: postgres
  ports:
    - name: postgres
      protocol: TCP
      port: 5432
      targetPort: 5432
```

Since we decided to use environment variables to configure the database connection of our operator we will have to add some default values in the `config/manager/manager.yaml` template. Here within the containers definition, we can define an `env` object that contains the necessary default values. Of course after changing this we also have to run the `make helm` command to reflect these changes in the helm chart's `values.yaml` file, which will be the place where the end-user would want to insert their own values.

```yaml
# config/manager/manager.yaml
env:
  - name: DATABASE_TYPE
    value: postgres
  - name: DSN
    value: postgres://user:password@postgres-service.default.svc.cluster.local:5432/db
```

![Database and WebSocket Setup](database-websocket-setup.png)

Finishing all these things we can confirm that the necessary dependencies are all installed and configured to work together so we can step over to the next phase of the implementation process.

## Backend Implementation

### The Main Function

At the start we have put some initializing logic into the `Reconcile()` method which is not the best place such cases since the `cmd/main.go` is the actual function that sets up the `ScannerReconciler` struct with the manager. Doing it this way the expressions do not have to be guarded by if statements of `sync.Once` constructs. Here we create a new `JobObjectService` connect to the database and pass the database reference to the newly created `ScanService`. The reconciler will need references of both services in order to list scanned images and launch new scans.

```go
// cmd/main.go within the main function

jobObjectService, err := service.NewJobObjectService()
if err != nil {
  mainLog.Error(err, "unable to create JobObjectService")
  os.Exit(1)
}

mainLog.Info("connecting to database")
db, err := database.GetDatabase()
if err != nil {
  mainLog.Error(err, "unable to connect to database")
  os.Exit(1)
}

scanService := service.NewScanService(db)

s := &http.Server{
  Handler: oapi.Handler(server.NewServer(scanService, mainLog)),
  Addr:    ":8000",
}

go func() {
  mainLog.Info("starting Scanner API HTTP server")
  if err := s.ListenAndServe(); err != http.ErrServerClosed {
    mainLog.Error(err, "unable to start Scanner API HTTP server")
    os.Exit(1)
  }
}()

// ...

if err = (&controller.ScannerReconciler{
  Client:           mgr.GetClient(),
  Scheme:           mgr.GetScheme(),
  JobObjectService: jobObjectService,
  ScanService:      scanService,
}).SetupWithManager(mgr); err != nil {
  mainLog.Error(err, "unable to create controller", "controller", "Scanner")
  os.Exit(1)
}
```

The `ScannerReconciler` struct is placed into the `internal/controller/scanner_controller.go` file, so the previously mentioned services have to be defined in this one. We can make use of interfaces to promote abstraction and allows us to achive loose coupling of different parts of our codebase. In accordance with Liskov substitution interfaces ensure that derived classes (or implementing types) can be substituted for the interface without affecting correctness, maintaining consistency in behavior. What this all means is that `ScannerReconciler` should not and does not have any insight into the inner workings of its dependencies.

```go
type ScannerReconciler struct {
  client.Client
  Scheme           *runtime.Scheme
  ScanService      service.ScanServiceInterface
  JobObjectService service.JobObjectServiceInterface
}
```

### Reconcile Loop

The reconciler runs on a single thread by default.

```go
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=list;watch;create
```

```go
func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
  reconcilerLog := log.FromContext(ctx)

  scanner := &scannerv1.Scanner{}
  if err := r.Get(ctx, req.NamespacedName, scanner); err != nil {
    reconcilerLog.Error(err, "unable to list scanner resources")
    return ctrl.Result{}, client.IgnoreNotFound(err)
  }

  scanResults, err := r.ScanService.ListScanResults()
  if err != nil {
    reconcilerLog.Error(err, "failed to list scan results")
    return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
  }
// ...
}
```

```go
func (r *ScannerReconciler) nextStatusCondition(
  ctx context.Context,
  scanner *scannerv1.Scanner,
  reason scannerv1.StatusReason,
) error {
  status := metav1.ConditionFalse
  if reason == scannerv1.Reconciled {
    status = metav1.ConditionTrue
  }

  changed := meta.SetStatusCondition(&scanner.Status.Conditions, metav1.Condition{
    Type:   "Ready",
    Status: status,
    Reason: string(reason),
  })

  if !changed {
    return nil
  }

  return r.Status().Update(ctx, scanner)
}
```

```go
  scannedImageIDs := []string{}
  for _, scanResult := range scanResults {
    scannedImageIDs = append(scannedImageIDs, scanResult.ImageID)
  }

  labelRequirement, err := labels.NewRequirement(scanner.Spec.IgnoreLabel, selection.NotEquals, []string{"true"})
  if err != nil {
    reconcilerLog.Error(err, "failed to get IgnoreLabel requirement")
    return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
  }

  podList := &corev1.PodList{}
  if err := r.List(ctx, podList, &client.ListOptions{
    Namespace:     scanner.Namespace,
    LabelSelector: labels.NewSelector().Add(*labelRequirement),
  }); err != nil {
    reconcilerLog.Error(err, "failed to list pods")
    return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
  }
```

```go
  imageID := ""
OuterLoop:
  for _, pod := range podList.Items {
    // TODO: Handle init containers as well
    for _, containerStatus := range pod.Status.ContainerStatuses {
      if !slices.Contains(scannedImageIDs, containerStatus.ImageID) {
        imageID = containerStatus.ImageID
        break OuterLoop
      }
    }
  }

  if imageID == "" {
    reconcilerLog.Info("all images scanned, successfully reconciled")
    return ctrl.Result{RequeueAfter: 10 * time.Second}, r.nextStatusCondition(ctx, scanner, scannerv1.Reconciled)
  }

```

```go
jobList := &batchv1.JobList{}
if err := r.List(ctx, jobList, &client.ListOptions{
  Namespace: scanner.Namespace,
}); err != nil {
  reconcilerLog.Error(err, "failed to list jobs")
  return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
}

for _, job := range jobList.Items {
  if job.Status.Succeeded == 0 {
    reconcilerLog.Info("job is still in progress")
    return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Waiting)
  }
}
```

```go
nextJob, err := r.JobObjectService.Create(imageID, scanner.Namespace)
if err != nil {
  reconcilerLog.Error(err, "failed to create job from template")
  return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
}

if err := ctrl.SetControllerReference(scanner, nextJob, r.Scheme); err != nil {
  reconcilerLog.Error(err, "failed to set controller reference on job")
  return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
}

if err := r.Create(ctx, nextJob); err != nil {
  reconcilerLog.Error(err, "failed to create job")
  return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Failed)
}

reconcilerLog.Info("new job created")
return ctrl.Result{}, r.nextStatusCondition(ctx, scanner, scannerv1.Scanning)
```

This allows a controller to ignore update events where the spec is unchanged, and only the metadata and/or status fields are changed.

```go
func (r *ScannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
  return ctrl.NewControllerManagedBy(mgr).
    For(&scannerv1.Scanner{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
    Owns(&batchv1.Job{}).
    Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.mapPodsToRequests)).
    Complete(r)
}
```

```go
func (r *ScannerReconciler) mapPodsToRequests(ctx context.Context, pod client.Object) []reconcile.Request {
  scannerList := &scannerv1.ScannerList{}
  if err := r.List(ctx, scannerList, &client.ListOptions{Namespace: pod.GetNamespace()}); err != nil {
    return []reconcile.Request{}
  }

  if len(scannerList.Items) > 0 {
    return []reconcile.Request{{NamespacedName: types.NamespacedName{
      Name:      scannerList.Items[0].Name,
      Namespace: scannerList.Items[0].Namespace,
    }}}
  }

  return []reconcile.Request{}
}
```

### Starting Jobs

- curl
- go templates

Here we are making use of the ttlSecondsAfterFinished property of jobs. It specifies the time-to-live (TTL) for the resource after it completes execution. Once the job finishes, Kubernetes waits for the defined TTL (in seconds) before automatically deleting the resource, helping to clean up old, unused jobs. This property is particularly useful for managing resource lifecycle and avoiding clutter in the cluster, but it must be enabled in the cluster settings as it is a feature gate. Using this takes the responsibility of deleting jobs away from our operator, but we can still have a look at what went wrong during the specified time. Once the unsucessful job is not present in the Namespace the Scanner operator schedule it again.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: { { .ScanName } }
  namespace: { { .Namespace } }
spec:
  ttlSecondsAfterFinished: 300
  backoffLimit: 0
  template:
    spec:
      initContainers:
        - name: grype
          image: anchore/grype:v0.83.0
          env:
            - name: GRYPE_DB_CACHE_DIR
              value: /grype-db
          args:
            - { { .ImageID } }
            - --output
            - cyclonedx-json
            - --file
            - /shared/scan-result.json
          volumeMounts:
            - name: shared
              mountPath: /shared
            - name: grype-db
              mountPath: /grype-db
      containers:
        - name: alpine
          image: alpine/curl:8.10.0
          command: ["sh", "-c"]
          args:
            - |
              echo '{"imageId":"{{.ImageID}}","report":'"$(cat /shared/scan-result.json)"'}\n' > /shared/scan-result.json;
              curl -X PUT -H 'Content-Type: application/json' -d @/shared/scan-result.json {{.ApiServiceHostname}}:8000/scan-results;
          volumeMounts:
            - name: shared
              mountPath: /shared
      restartPolicy: Never
      volumes:
        - name: shared
          emptyDir: {}
        - name: grype-db
          hostPath:
            path: /grype-db
```

TODO: why

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: /home/outis/.cache/grype/db
        containerPath: /grype-db
```

### HTTP Handlers

TODO: defer

```go
func (s *Server) PutScanResults(w http.ResponseWriter, r *http.Request) {
  defer observeDuration("PUT", "/scan-results")()
  oapiScanResult := oapi.ScanResult{}
  if err := json.NewDecoder(r.Body).Decode(&oapiScanResult); err != nil {
    s.logger.Error(err, "PutScanResults")
    http.Error(w, "Bad Request", http.StatusBadRequest)
    return
  }

  scanResult, err := s.scanService.UpsertScanResult(oapiScanResult.ImageId, string(oapiScanResult.Report))
  if errors.Is(err, service.InvalidCycloneDXBOM) {
    s.logger.Error(err, "PutScanResults")
    http.Error(w, "Bad Request", http.StatusBadRequest)
    return
  } else if err != nil {
    s.logger.Error(err, "PutScanResults")
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    return
  }

  s.broadcastCh <- scanResult.ImageID
  s.logger.Info("PutScanResults", "new imageId broadcasted", scanResult.ImageID)

  res := oapi.ScanResult{
    ImageId: scanResult.ImageID,
    Report:  json.RawMessage(scanResult.Report),
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(res); err != nil {
    s.logger.Error(err, "PutScanResults")
  }
}
```

TODO: mutex
TODO: sync once
TODO: channels
TODO: prometheus

```go
func (s *Server) GetSubscribe(w http.ResponseWriter, r *http.Request) {
  defer observeDuration("GET", "/subscribe")()
  c, err := s.upgrader.Upgrade(w, r, nil)
  if err != nil {
    s.logger.Error(err, "GetSubscribe")
    return
  }

  s.once.Do(func() {
    go func() {
      for {
        message := <-s.broadcastCh

        for _, ch := range s.connections {
          ch <- message
        }
      }
    }()
  })

  defer c.Close()

  ch := make(chan string)
  defer close(ch)

  s.mu.Lock()
  s.connections[c] = ch
  s.mu.Unlock()

  go func() {
    for {
      imageId, ok := <-ch
      if !ok {
        return
      }

      data := []byte("\"" + imageId + "\"")
      if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
        s.logger.Error(err, "Websocket")
      }
    }
  }()

  for {
    msgType, _, err := c.ReadMessage()
    if err != nil {
      break
    }

    if msgType == websocket.CloseMessage {
      break
    }
  }

  s.mu.Lock()
  delete(s.connections, c)
  s.mu.Unlock()
}
```

### Scan Service and Database Model

```go
type ScanResult struct {
  ImageID string `gorm:"primarykey;type:TEXT"`
  Report  string `gorm:"not null;type:TEXT"`
}
```

```go
func (s *ScanService) UpsertScanResult(imageId string, report string) (*database.ScanResult, error) {
  bom := cyclonedx.BOM{}
  reader := strings.NewReader(report)
  decoder := cyclonedx.NewBOMDecoder(reader, cyclonedx.BOMFileFormatJSON)
  if err := decoder.Decode(&bom); err != nil {
    return nil, fmt.Errorf("%w: %w", InvalidCycloneDXBOM, err)
  }

  scanResult := database.ScanResult{
    ImageID: imageId,
    Report:  report,
  }

  res := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&scanResult)
  if res.Error != nil {
    return nil, fmt.Errorf("error while inserting ScanResult: %w", res.Error)
  }

  return &scanResult, nil
}
```

CRUD stands for an API that supports creating, retrieving, updating and deleting a certain resource.

For simplicity, we can use the TEXT datatype which can store a near unlimited number of bytes and is available in postgres, mysql and sqlite too, so it is suitable to store manifests.

### Prometheus

Prometheus is an open-source monitoring and alerting tool designed for collecting, storing, and querying metrics from various systems and applications. It uses a time-series database to store metrics data, allowing for analysis of trends and real-time monitoring. We can define a make command so installing [kube-prometheus](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) can be done automatically.

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

```Makefile
.PHONY: setup-prometheus
setup-prometheus:
  kubectl create namespace monitoring
  helm install -n monitoring prometheus prometheus-community/kube-prometheus-stack
  kubectl port-forward service/prometheus-operated -n monitoring 9090:9090
```

Prometheus scrapes metrics from target systems using HTTP endpoints, typically in a specific text-based format. It features a powerful query language called PromQL for data analysis and visualization, often integrated with dashboards like Grafana. Prometheus also includes built-in alerting capabilities, enabling users to define conditions and receive notifications when thresholds are breached.

Enabling Prometheus metrics is as easy as uncommenting the proper section in `config/default/kustomization.yaml` and also add a tag to the already defined `ServiceMonitor` resource so that Prometheus operator can register it. Of course after any modification involving the `config` directory we should regenerate the helm chart.

## Frontend Implementation

### Custom Hook

We can create a proper wrapper hook for the `Subscriber` class to make it work well within React components. A React hook is a special function that lets you use React features in functional components, like state and lifecycle methods. Preexisting hooks, such as `useState` and `useEffect`, simplify component logic and enable reusability of stateful behavior across components and are the building blocks that enable us to build more complex components. The convention is to start the name of such function with the 'use' word in order to signal other developers about its nature.

Our the `useSubscriber` hook is a modified version of the previously showed wrapper around the Subscriber class. It uses the `Subscriber` class to listen to events from a given URL ("/subscribe"). The hook accepts two callback props, `onMessage` and `onConnection`, which are called when respective "message" and "connection" events are triggered, passing along relevant data (e.detail) from each event. Inside the effect, once again an AbortController (ac) is created to allow cancelling and proper termination of the subscription when the component unmounts from the virtual DOM, by calling `ac.abort()` in the cleanup function. The `onMessage` and `onConnection` handlers are passed into the dependency list of `useEffect`, so that the hook is refreshed when the handlers change.

```tsx
// frontend/src/hooks/useSubscriber.tsx
export function useSubscriber({ onMessage, onConnection }: UseSubscriberProps) {
  useEffect(() => {
    const ac = new AbortController();
    const s = new Subscriber("/subscribe", { signal: ac.signal });

    s.addEventListener("message", (e: CustomEventInit) => {
      onMessage(e.detail as string);
    });

    s.addEventListener("connection", (e: CustomEventInit) => {
      onConnection(e.detail as boolean);
    });

    return () => ac.abort();
  }, [onMessage, onConnection]);
}
```

### State Management

With the help of `createContext` function we can implement a global state manager which holds and makes available the application's state and a `dispatch` function for state updates for other components. The `GlobalStateProvider` component employs the `useReducer` hook to manage the application's state using `globalReducer`, starting with initialState. Within this provider, two callback functions, `onMessage` and `onConnection`, handle specific events:

- `onMessage` dispatches an "add" action to add a new item based on incoming messages
- `onConnection` fetches all scan result objects from the backend and dispatches a "connection_gained" with the payload when the `Subscriber` instance transitions from offline to online state and a "connection_lost" action when the reverse happens. This way of operation ensures that when online all data shown on the UI is representing the current state of the cluster.

These two functions have to be defined using the `useCallback` hooks to cache them between rerenders, otherwise the `Subscriber` class would be detatched and re-attached repeatedly in an uncontrolled manner. `useSubscriber` hooks into these callbacks, and the provider component passes the context value, allowing child components to access and interact with the global state.

Here we have to use the `encodeURIComponent()` function in order to escape some special characters like forward slash and the at sign, since the imageId contains such characters.

```tsx
// frontend/src/components/GlobalState.tsx
export const GlobalStateContext = createContext({
  state: initialState,
  dispatch: (action: Action) => {},
});

export type GlobalStateProviderProps = PropsWithChildren<{}>;

export const GlobalStateProvider = ({ children }: GlobalStateProviderProps) => {
  const [state, dispatch] = useReducer(globalReducer, initialState);

  const onMessage = useCallback(
    async (imageId: string) => {
      const res = await fetch(`/scan-results/${encodeURIComponent(imageId)}`);
      if (!res.ok) {
        return;
      }

      const scanResult = (await res.json()) as ScanResult;
      dispatch({ type: "add", payload: scanResult });
    },
    [dispatch]
  );

  const onConnection = useCallback(
    async (isConnected: boolean) => {
      if (isConnected) {
        const res = await fetch(`/scan-results`);
        if (!res.ok) {
          return;
        }

        const scanResults = (await res.json()) as ScanResult[];
        dispatch({ type: "connection_gained", payload: scanResults });
      } else {
        dispatch({ type: "connection_lost" });
      }
    },
    [dispatch]
  );

  useSubscriber({ onMessage, onConnection });

  return (
    <GlobalStateContext.Provider value={{ state, dispatch }}>
      {children}
    </GlobalStateContext.Provider>
  );
};
```

The initialState includes two properties: `isConnected`, a boolean for connection status, and `scanResults`, an array of scan results. The Action type defines four possible actions: "add" to add a new scan result, "remove" to delete a scan result by id, and "connection_gained" with a payload that contains all the currently available scan results and "connection_lost" to show that the subscriber does not have a live connection. The `globalReducer` function derives the next state based on the action type. This setup allows for flexible and predictable state management, supporting multiple actions that modify the global state in specific ways so the mutation logic is separated from visualization. The `ScanResult` type is imported from the `oapi.gen.d.ts` which was generated based upon the OpenAPI schema by `openapi-typescript`.

```tsx
// frontend/src/components/GlobalState.tsx
const initialState = {
  isConnected: false,
  scanResults: [] as ScanResult[],
};
type State = typeof initialState;
type Action =
  | { type: "add"; payload: ScanResult }
  | { type: "remove"; payload: ScanResult["imageId"] }
  | { type: "connection_gained"; payload: ScanResult[] }
  | { type: "connection_lost" };

function globalReducer(state: State, action: Action): State {
  switch (action.type) {
    case "connection_lost": {
      return { ...state, isConnected: false };
    }
    case "connection_gained": {
      return {
        ...state,
        isConnected: true,
        scanResults: action.payload.sort((a, b) =>
          a.imageId.localeCompare(b.imageId)
        ),
      };
    }
    case "add": {
      const index = state.scanResults.findIndex(
        (item) => item.imageId === action.payload.imageId
      );

      let scanResults: ScanResult[] = [];
      if (index !== -1) {
        scanResults = state.scanResults.map((s, i) =>
          i === index ? action.payload : s
        );
      } else {
        scanResults = [...state.scanResults, action.payload];
      }

      return {
        ...state,
        scanResults: scanResults.sort((a, b) =>
          a.imageId.localeCompare(b.imageId)
        ),
      };
    }
    case "remove": {
      return {
        ...state,
        scanResults: state.scanResults.filter(
          (s) => s.imageId !== action.payload
        ),
      };
    }
  }
}
```

So accessing and updating global state from within a component is as simple as passing the `GlobalStateContext` into `useContext` within our component. This is exacly what happens within the `ScanResultList` component.

```tsx
const { state, dispatch } = useContext(GlobalStateContext);
```

The following image shows the finished UI, where the scanner resource was installed in the default namespace and it is in a reconciled state.

![Finished UI](finished-ui.png)

## Testing and Monitoring using Prometheus

### Manual Flow

During development we can make our job a lot more easier by automating what we can. Makefiles can be used for such purposes too, so we defined a `make cycle` commands which tears down the cluster if it was set up, then recreates it, builds frontend code, starts the postgres database, builds the docker image of the operator, loads the newly buildt image into the kind cluster, deploys the helm chart and applies a scanner resource in the default namespace.

```makefile
.PHONY: cycle
cycle: teardown-cluster setup-cluster build-frontend setup-db docker-build kind-load helm-deploy start-scanner
```

TODO: how did this appear

```sh
$ kubectl get all
NAME                            READY   STATUS     RESTARTS   AGE
pod/postgres-786b469c56-wzx89   1/1     Running    0          99s
pod/scan-iegy0oylsg-xxl9v       0/1     Init:0/1   0          20s

NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/kubernetes         ClusterIP   10.96.0.1      <none>        443/TCP    108s
service/postgres-service   ClusterIP   10.96.40.196   <none>        5432/TCP   99s

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/postgres   1/1     1            1           99s

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/postgres-786b469c56   1         1         1       99s

NAME                        STATUS    COMPLETIONS   DURATION   AGE
job.batch/scan-iegy0oylsg   Running   0/1           20s        20s

NAME                                               IGNORELABEL
scanner.scanner.zoltankerezsi.xyz/scanner-sample   ignore
```

To test the whole system, we can start a few pods and wait until they automatically appear on the UI.

```sh
kubectl run python-sleep --image=python:latest --restart=Never -- sleep 3600
kubectl run alpine-sleep --image=alpine:latest --restart=Never -- sleep 3600
```

We can label a specific pod with the `ignore=true`. By clicking the `Delete` button on one of the items it should dissapear from the list, and not be rescheduled again for scanning.

```sh
kubectl label pod alpine-sleep ignore=true
```

TODO: describe response image

![Prometheus](prometheus.png)

TODO: list scan resource. reconciled true

With all these tested and working, we can conclude that we are done developing the operator, all three aspect it (controller-manager, backend, and frontend) is functioning correctly in themselves and together.

## Other Resources

- <https://maelvls.dev/kubernetes-conditions/>
- <https://book.kubebuilder.io/reference/metrics>
- <https://prometheus.io/docs/guides/go-application/>
