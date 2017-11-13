package google

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func resourceGoogleOrganizationIamPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceGoogleOrganizationIamPolicyCreate,
		Read:   resourceGoogleOrganizationIamPolicyRead,
		Update: resourceGoogleOrganizationIamPolicyUpdate,
		Delete: resourceGoogleOrganizationIamPolicyDelete,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_data": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: jsonPolicyDiffSuppress,
				ValidateFunc:     validateIamPolicy,
			},
		},
	}
}

func resourceGoogleOrganizationIamPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	if err := setOrganizationIamPolicy(d, config); err != nil {
		return fmt.Errorf("Error attaching the organization iam policy: %s", err)
	}

	d.SetId("organizations/" + d.Get("org_id").(string))

	return resourceGoogleOrganizationIamPolicyRead(d, meta)
}

func resourceGoogleOrganizationIamPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	policy, err := config.clientResourceManager.Organizations.GetIamPolicy(d.Id(), &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Iam policy for %s", d.Id()))
	}

	d.Set("policy_data", marshalIamPolicy(policy))

	return nil
}

func resourceGoogleOrganizationIamPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	if d.HasChange("policy_data") {
		if err := setOrganizationIamPolicy(d, config); err != nil {
			return fmt.Errorf("Error updating the organization policy: %s", err)
		}
	}

	return resourceGoogleOrganizationIamPolicyRead(d, meta)
}

func resourceGoogleOrganizationIamPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	_, err := config.clientResourceManager.Organizations.SetIamPolicy(d.Id(), &cloudresourcemanager.SetIamPolicyRequest{
		Policy:     &cloudresourcemanager.Policy{},
		UpdateMask: "bindings",
	}).Do()

	if err != nil {
		return fmt.Errorf("Error deleting the organization policy: %s", err)
	}

	return nil
}

func setOrganizationIamPolicy(d *schema.ResourceData, config *Config) error {
	org := "organizations/" + d.Get("org_id").(string)
	policy, err := unmarshalIamPolicy(d.Get("policy_data").(string))
	if err != nil {
		return fmt.Errorf("'policy_data' is not valid for %s: %s", org, err)
	}

	_, err = config.clientResourceManager.Organizations.SetIamPolicy(org, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()

	return err
}

func marshalIamPolicy(policy *cloudresourcemanager.Policy) string {
	pdBytes, _ := json.Marshal(&cloudresourcemanager.Policy{
		Bindings: policy.Bindings,
	})
	return string(pdBytes)
}

func unmarshalIamPolicy(policyData string) (*cloudresourcemanager.Policy, error) {
	policy := &cloudresourcemanager.Policy{}
	if err := json.Unmarshal([]byte(policyData), policy); err != nil {
		return nil, fmt.Errorf("Could not unmarshal policy data %s:\n%s", policyData, err)
	}
	return policy, nil
}

func validateIamPolicy(i interface{}, k string) (s []string, es []error) {
	_, err := unmarshalIamPolicy(i.(string))
	if err != nil {
		es = append(es, err)
	}
	return
}
