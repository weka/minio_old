/*
 * MinIO Cloud Storage, (C) 2016, 2017, 2018, 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	wekaPathPrefix         = minioReservedBucketPath + "/weka"
	wekaAPIVersion         = "v1"
	wekaAPIPrefix  	  = slashSeparator + wekaAPIVersion
)

// WekaAPIHandlers provides HTTP handlers for MinIO admin API.
type WekaAPIHandlers struct{}

func (h WekaAPIHandlers) RefreshList(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "StartProfiling")

	_ = globalIAMSys.store.loadAll(ctx, globalIAMSys)
}

// registerWekaRouter - Add handler functions for each service REST API routes.
func registerWekaRouter(router *mux.Router) {

	wekaAPI := WekaAPIHandlers{}
	// Admin router
	wekaRouter := router.PathPrefix(wekaPathPrefix).Subrouter()

	/// Service operations
	wekaRouter.Methods(http.MethodPost).Path("/refresh").HandlerFunc(httpTraceAll(wekaAPI.RefreshList))
	wekaRouter.Methods(http.MethodGet).Path("/refresh").HandlerFunc(httpTraceAll(wekaAPI.RefreshList))


	// If none of the routes match add default error handler routes
	//wekaRouter.NotFoundHandler = httpTraceAll(errorResponseHandler)
	//wekaRouter.MethodNotAllowedHandler = httpTraceAll(methodNotAllowedHandler("Weka"))
}
