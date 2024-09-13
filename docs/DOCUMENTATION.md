# Scanner Operator

## Overview

In this article I would like to introduce the reader into the realms of cloud native technologies by showing how one can implement a fully custom operator using Go and Kubebuilder. While I don't consider myself an expert in this field during my internship at Cisco I think acquired enough knowledge to educate others on this matter and hopefully dissolve some misconceptions about its difficulty.

In order to make this guide more similar the process one would do in reality when meaningful work is needed I would like to concentrate on things that go a bit further than simply achieving a "hello world" type of result.

## Kubernetes and containerization

## Initial setup of the project

First we have to make sure we have the most recent stable version of go and kubectl CLI tools.
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

Then we download the latest release of Kubebuilder CLI, make it executable and move it into out `/usr/local/bin` where user installed binaries are usually stored.

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
We can make these commands callable by putting this directory onto our $PATH if needed.

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

After these steps we will have a scaffold generated in the `internal/controller` folder for us. The two most important methods are the `Reconciler` and the `SetupWithManager` which we will take a more in depth look at later. Also in order to setup a cluster we can use a tool called [kind](https://kind.sigs.k8s.io/) which makes it easy to create a delete Kubernetes clusters and nodes for our development.

```sh
go install sigs.k8s.io/kind@v0.24.0
kind create cluster
```

We can verify if kubectl is configured correctly for this cluster.

```sh
$ kubectl cluster-info --context kind-kind
Kubernetes control plane is running at https://127.0.0.1:37675
CoreDNS is running at https://127.0.0.1:37675/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

## Setting up the UI development environment

The UI that we assemble here is considered as the test or proof that the operator does what it has to with a reasonable performance.

For simplicitys sake, I choosed to develop this UI using React, since the tooling around it is very mature. Also, we will be using Node instead of the newer more modern javascript runtimes like Bun or Deno. These are functionally mostly compatible with Node but there could still be some rough edges or surprizing hardships when you are trying to achieve an exact result.

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

- `esbuild` is a fast bundler for Javascript and Typescript which is a superset of Javascript. A bundler is a tool that takes multiple source code files and combines them into one or more depending on the configuration. Before the `import` directive was available, using a bundler was our only choice to ship code with multiple dependencies if we did not want to use global scoped objects as with [JQuery](https://jquery.com/). Now that [module syntax](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules) is standardized, it is still useful to bundle our frontend code to minimize the number of network requests the browser has to make in order to gather all of the source files and run our code. The other reason is of course that we use Typescript which is directly not runnable by the browser, so a tranformation step is necessary.
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

root.render(App());
```

We also define some script in our `package.json` file to document the steps it takes to build the final javascript and CSS files which can later be copied and served using a HTTP server. Here we are using the "production" `NODE_ENV` enviroment variable which instructs the bundler to omit program code that would enable us to attach certain debugging tools like the React Developer Tools that makes browsers aware of reacts internal state and behavior. This value can be changed anytime.

```json
{
    "scripts": {
        "build-css": "./node_modules/.bin/tailwindcss -i input.css -o output.css",
        "build-js": "./node_modules/.bin/esbuild index.tsx --define:process.env.NODE_ENV=\\\"production\\\" --bundle --outfile=bundle.js",
        "build": "npm run build-css && npm run build-js"
    },
    ...
}
```

## Resources

<https://esbuild.github.io/>