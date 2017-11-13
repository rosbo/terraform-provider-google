package google

import (
	"testing"
	"google.golang.org/api/cloudresourcemanager/v1"
	"github.com/hashicorp/terraform/helper/resource"
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGoogleOrganizationIamPolicy_basic(t *testing.T) {
	t.Parallel()

	accountId := acctest.RandomWithPrefix("tf-test")

	skipIfEnvNotSet(t, "GOOGLE_ORG")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGoogleOrganizationIamPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccGoogleOrganizationIamPolicy(org, accountId, "roles/viewer"),
				Check: testAccCheckGoogleOrganizationIamPolicy("google_organization_iam_policy.test", "roles/viewer"),
			},
			resource.TestStep{
				Config: testAccGoogleOrganizationIamPolicy(org, accountId, "roles/editor"),
				Check: testAccCheckGoogleOrganizationIamPolicy("google_organization_iam_policy.test", "roles/editor"),
			},
		},
	})
}

func testAccCheckGoogleOrganizationIamPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_organization_iam_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, err := config.clientResourceManager.Organizations.GetIamPolicy(rs.Primary.ID, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
		if err != nil && len(policy.Bindings) > 0 {
			return fmt.Errorf("Organization policy for %q hasn't been deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckGoogleOrganizationIamPolicy(n, role string) resource.TestCheckFunc {

}

func testAccGoogleOrganizationIamPolicy(orgId, accountId, role string) string {
	return fmt.Sprintf(`
resource "google_service_account" "test-account" {
	account_id = "%s"
}

data "google_iam_policy" "test" {
  binding {
    role = "%s"

    members = ["${google_service_account.test-account.email}"]
  }
}

resource "google_organization_iam_policy" "test" {
  org_id      = "%s"
  policy_data = "${data.google_iam_policy.test.policy_data}"
}
`, accountId, role, orgId)
}
