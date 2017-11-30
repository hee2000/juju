// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package caasoperator

import (
	"github.com/juju/errors"
	"gopkg.in/juju/charm.v6"
	"gopkg.in/juju/names.v2"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/api/common"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/status"
	"github.com/juju/juju/watcher"
)

// Client allows access to the CAAS operator API endpoint.
type Client struct {
	facade base.FacadeCaller
}

// NewClient returns a client used to access the CAAS Operator API.
func NewClient(caller base.APICaller) *Client {
	facadeCaller := base.NewFacadeCaller(caller, "CAASOperator")
	return &Client{
		facade: facadeCaller,
	}
}

func (c *Client) tag(application string) (names.ApplicationTag, error) {
	if !names.IsValidApplication(application) {
		return names.ApplicationTag{}, errors.NotValidf("application name %q", application)
	}
	return names.NewApplicationTag(application), nil
}

// SetStatus sets the status of the specified appplication.
func (c *Client) SetStatus(
	application string,
	status status.Status,
	info string,
	data map[string]interface{},
) error {
	tag, err := c.tag(application)
	if err != nil {
		return errors.Trace(err)
	}
	var result params.ErrorResults
	args := params.SetStatus{
		Entities: []params.EntityStatusArgs{{
			Tag:    tag.String(),
			Status: status.String(),
			Info:   info,
			Data:   data,
		}},
	}
	if err := c.facade.FacadeCall("SetStatus", args, &result); err != nil {
		return errors.Trace(err)
	}
	return result.OneError()
}

// Charm returns information about the charm currently assigned
// to the application.
func (c *Client) Charm(application string) (_ *charm.URL, sha256 string, _ error) {
	tag, err := c.tag(application)
	if err != nil {
		return nil, "", errors.Trace(err)
	}
	var results params.ApplicationCharmResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: tag.String()}},
	}
	if err := c.facade.FacadeCall("Charm", args, &results); err != nil {
		return nil, "", errors.Trace(err)
	}
	if n := len(results.Results); n != 1 {
		return nil, "", errors.Errorf("expected 1 result, got %d", n)
	}
	if err := results.Results[0].Error; err != nil {
		return nil, "", errors.Trace(err)
	}
	result := results.Results[0].Result
	curl, err := charm.ParseURL(result.URL)
	if err != nil {
		return nil, "", errors.Trace(err)
	}
	return curl, result.SHA256, nil
}

// WatchApplicationConfig returns a watcher that is notified whenever the
// application config changes.
func (c *Client) WatchApplicationConfig(application string) (watcher.NotifyWatcher, error) {
	tag, err := c.tag(application)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return common.Watch(c.facade, "WatchApplicationConfig", tag)
}

// ApplicationConfig returns the application's config settings.
func (c *Client) ApplicationConfig(application string) (charm.Settings, error) {
	tag, err := c.tag(application)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var results params.ConfigSettingsResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: tag.String()}},
	}
	if err := c.facade.FacadeCall("ApplicationConfig", args, &results); err != nil {
		return nil, errors.Trace(err)
	}
	if len(results.Results) != 1 {
		return nil, errors.Errorf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if result.Error != nil {
		return nil, errors.Trace(result.Error)
	}
	return charm.Settings(result.Settings), nil
}
