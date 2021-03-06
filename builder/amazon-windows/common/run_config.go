package common

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/packer/packer"
)

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	AssociatePublicIpAddress bool              `mapstructure:"associate_public_ip_address"`
	AvailabilityZone         string            `mapstructure:"availability_zone"`
	IamInstanceProfile       string            `mapstructure:"iam_instance_profile"`
	InstanceType             string            `mapstructure:"instance_type"`
	RunTags                  map[string]string `mapstructure:"run_tags"`
	SourceAmi                string            `mapstructure:"source_ami"`
	SpotPrice                string            `mapstructure:"spot_price"`
	SpotPriceAutoProduct     string            `mapstructure:"spot_price_auto_product"`
	WinRMPrivateIp           bool              `mapstructure:"winrm_private_ip"`
	SecurityGroupId          string            `mapstructure:"security_group_id"`
	SecurityGroupIds         []string          `mapstructure:"security_group_ids"`
	SubnetId                 string            `mapstructure:"subnet_id"`
	UserData                 string            `mapstructure:"user_data"`
	UserDataFile             string            `mapstructure:"user_data_file"`
	VpcId                    string            `mapstructure:"vpc_id"`
}

func (c *RunConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	templates := map[string]*string{
		"iam_instance_profile":    &c.IamInstanceProfile,
		"instance_type":           &c.InstanceType,
		"spot_price":              &c.SpotPrice,
		"spot_price_auto_product": &c.SpotPriceAutoProduct,
		"source_ami":              &c.SourceAmi,
		"subnet_id":               &c.SubnetId,
		"vpc_id":                  &c.VpcId,
		"availability_zone":       &c.AvailabilityZone,
		"user_data":               &c.UserData,
		"user_data_file":          &c.UserDataFile,
		"security_group_id":       &c.SecurityGroupId,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	// Validation
	if c.SourceAmi == "" {
		errs = append(errs, errors.New("A source_ami must be specified"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An instance_type must be specified"))
	}

	if c.SpotPrice == "auto" {
		if c.SpotPriceAutoProduct == "" {
			errs = append(errs, errors.New(
				"spot_price_auto_product must be specified when spot_price is auto"))
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.SecurityGroupId != "" {
		if len(c.SecurityGroupIds) > 0 {
			errs = append(errs, fmt.Errorf("Only one of security_group_id or security_group_ids can be specified."))
		} else {
			c.SecurityGroupIds = []string{c.SecurityGroupId}
			c.SecurityGroupId = ""
		}
	}

	sliceTemplates := map[string][]string{
		"security_group_ids": c.SecurityGroupIds,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = t.Process(elem, nil)
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	newTags := make(map[string]string)
	for k, v := range c.RunTags {
		k, err := t.Process(k, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing tag key %s: %s", k, err))
			continue
		}

		v, err := t.Process(v, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing tag value '%s': %s", v, err))
			continue
		}

		newTags[k] = v
	}

	c.RunTags = newTags

	return errs
}
