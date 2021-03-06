package cloudflare

import (
	"fmt"
	"strconv"
	"testing"

	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccCloudflareLogpushJob_Basic(t *testing.T) {
	jobID := acctest.RandString(10)
	name := "cloudflare_logpush_job." + jobID
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	ownershipToken := os.Getenv("CLOUDFLARE_LOGPUSH_OWNERSHIP_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckLogpushToken(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudflareLogpushJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareLogpushJobConfig(jobID, zoneID, ownershipToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "test-logpush-job"),
					resource.TestCheckResourceAttr(name, "enabled", "true"),
					resource.TestCheckResourceAttr(name, "logpull_options", "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339"),
					resource.TestCheckResourceAttr(name, "destination_conf", "s3://logpush-test-bucket?region=us-west-1"),
				),
			},
		},
	})
}

func testAccCheckCloudflareLogpushJobDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_logpush_job" {
			continue
		}

		primaryID, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Could not retrieve LogpushJob ID")
		}

		_, fetchErr := client.LogpushJob(rs.Primary.Attributes["zone_id"], primaryID)
		if fetchErr == nil {
			return fmt.Errorf("Logpush job still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudflareLogpushJobConfig(jobID, zoneID, ownershipToken string) string {
	return fmt.Sprintf(`
resource "cloudflare_logpush_job" "%[1]s" {
  name = "test-logpush-job"
  enabled = "true" 
  zone_id = "%[2]s"
  destination_conf = "s3://logpush-test-bucket?region=us-west-1"
  logpull_options = "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339"
  ownership_challenge = "%[3]s"
}`, jobID, zoneID, ownershipToken)
}
