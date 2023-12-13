// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitReports() {
	api.BaseRoutes.Reports.Handle("/users", api.APISessionRequired(getUsersForReporting)).Methods("GET")
}

func getUsersForReporting(c *Context, w http.ResponseWriter, r *http.Request) {
	if !(c.IsSystemAdmin() && c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadUserManagementUsers)) {
		c.SetPermissionError(model.PermissionSysconsoleReadUserManagementUsers)
		return
	}

	sortColumn := "Username"
	if r.URL.Query().Get("sort_column") != "" {
		sortColumn = r.URL.Query().Get("sort_column")
	}

	pageSize := 50
	if pageSizeStr, err := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 64); err == nil {
		pageSize = int(pageSizeStr)
	}

	teamFilter := r.URL.Query().Get("team_filter")
	if !(teamFilter == "" || model.IsValidId(teamFilter)) {
		c.Err = model.NewAppError("getUsersForReporting", "api.getUsersForReporting.invalid_team_filter", nil, "", http.StatusBadRequest)
		return
	}

	hideActive := r.URL.Query().Get("hide_active") == "true"
	hideInactive := r.URL.Query().Get("hide_inactive") == "true"
	if hideActive && hideInactive {
		c.Err = model.NewAppError("getUsersForReporting", "api.getUsersForReporting.invalid_active_filter", nil, "", http.StatusBadRequest)
		return
	}

	options := &model.UserReportOptions{
		ReportingBaseOptions: model.ReportingBaseOptions{
			SortColumn:          sortColumn,
			SortDesc:            r.URL.Query().Get("sort_direction") == "desc",
			PageSize:            pageSize,
			LastSortColumnValue: r.URL.Query().Get("last_column_value"),
			DateRange:           r.URL.Query().Get("date_range"),
		},
		Team:         teamFilter,
		LastUserId:   r.URL.Query().Get("last_id"),
		Role:         r.URL.Query().Get("role_filter"),
		HasNoTeam:    r.URL.Query().Get("has_no_team") == "true",
		HideActive:   hideActive,
		HideInactive: hideInactive,
	}
	options.PopulateDateRange(time.Now())

	userReports, err := c.App.GetUsersForReporting(options)
	if err != nil {
		c.Err = err
		return
	}

	if jsonErr := json.NewEncoder(w).Encode(userReports); jsonErr != nil {
		c.Logger.Warn("Error writing response", mlog.Err(jsonErr))
	}
}
