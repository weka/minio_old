/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
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
	"fmt"
	"net/http"
	"strconv"
)

func drainModePutHandler(w http.ResponseWriter, r *http.Request) {
	drainMode, _ := strconv.ParseBool(r.URL.Query().Get("value"))
	globalDrainMode = drainMode
	writeResponse(w, http.StatusOK, []byte(fmt.Sprintf("{'Mode':'%t'}", globalDrainMode)), mimeNone)
}

func drainModeGetHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, []byte(fmt.Sprintf("{'Mode':'%t'}", globalDrainMode)), mimeNone)
}

func drainStatusGetHandler(w http.ResponseWriter, r *http.Request) {
	type DrainStatus struct {
		Mode string `json:"mode"`
		Status string `json:"status"`
	}

	status := DrainStatus{
		Mode: strconv.FormatBool(globalDrainMode),
		Status: strconv.FormatInt(globalDrainStatus, 10),
	}

	writeResponse(w, http.StatusOK, encodeResponseJSON(status), mimeNone)
}
