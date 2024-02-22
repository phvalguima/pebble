// Copyright (c) 2022 Canonical Ltd
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License version 3 as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package daemon

import (
	"net/http"
	"time"

	"github.com/canonical/x-go/strutil"

	"github.com/canonical/pebble/internals/logger"
	"github.com/canonical/pebble/internals/overlord/checkstate"
	"github.com/canonical/pebble/internals/plan"
)

type healthInfo struct {
	Healthy bool `json:"healthy"`
}

func v1Health(c *Command, r *http.Request, _ *UserState) Response {
	start_ts := time.Now()
	logger.Noticef("api_health.go v1Health from %s started at %s", r.RemoteAddr, start_ts.String())

	query := r.URL.Query()
	level := plan.CheckLevel(query.Get("level"))
	switch level {
	case plan.UnsetLevel, plan.AliveLevel, plan.ReadyLevel:
	default:
		return BadRequest(`level must be "alive" or "ready"`)
	}
	logger.Noticef("api_health.go v1Health from %s level calculation: %v", r.RemoteAddr, time.Since(start_ts))

	names := strutil.MultiCommaSeparatedList(query["names"])
	logger.Noticef("api_health.go v1Health from %s MultiCommaSeparatedList: %v", r.RemoteAddr, time.Since(start_ts))

	checks, err := getChecks(c.d.overlord)
	if err != nil {
		logger.Noticef("Cannot fetch checks: %v", err.Error())
		return InternalError("internal server error")
	}
	logger.Noticef("api_health.go v1Health from %s getChecks: %v", r.RemoteAddr, time.Since(start_ts))

	healthy := true
	status := http.StatusOK
	for _, check := range checks {
		levelMatch := level == plan.UnsetLevel || level == check.Level ||
			level == plan.ReadyLevel && check.Level == plan.AliveLevel // ready implies alive
		namesMatch := len(names) == 0 || strutil.ListContains(names, check.Name)
		if levelMatch && namesMatch && check.Status != checkstate.CheckStatusUp {
			healthy = false
			status = http.StatusBadGateway
		}
	}
	logger.Noticef("api_health.go v1Health from %s level processing: %v", r.RemoteAddr, time.Since(start_ts))

	sync_resp := SyncResponse(&resp{
		Type:   ResponseTypeSync,
		Status: status,
		Result: healthInfo{Healthy: healthy},
	})
	logger.Noticef("api_health.go v1Health from %s sync response: %v", r.RemoteAddr, time.Since(start_ts))
	logger.Noticef("api_health.go v1Health from %s finished: %s", r.RemoteAddr, time.Now().String())
	return sync_resp
}
