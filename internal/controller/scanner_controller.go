/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"net/http"
	"os"

	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	scannerv1 "github.com/kerezsiz42/scanner-operator2/api/v1"
	"github.com/kerezsiz42/scanner-operator2/internal/database"
	"github.com/kerezsiz42/scanner-operator2/internal/oapi"
	"github.com/kerezsiz42/scanner-operator2/internal/server"
)

// ScannerReconciler reconciles a Scanner object
type ScannerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Server *http.Server
	Db     *gorm.DB
}

// +kubebuilder:rbac:groups=scanner.zoltankerezsi.xyz,resources=scanners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scanner.zoltankerezsi.xyz,resources=scanners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scanner.zoltankerezsi.xyz,resources=scanners/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scanner object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
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

// SetupWithManager sets up the controller with the Manager.
func (r *ScannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scannerv1.Scanner{}).
		Complete(r)
}
