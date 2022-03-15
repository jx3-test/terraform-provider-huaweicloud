package ecs

import (
	"fmt"
	"testing"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/ecs"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk/openstack/ecs/v1/nics"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
)

func getEcsNictResourceFunc(config *config.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := config.ComputeV1Client(acceptance.HW_REGION_NAME)
	if err != nil {
		return nil, fmtp.Errorf("Error creating Compute v1 client, err=%s", err)
	}

	instanceId, portId, err := ecs.ComputeInterfaceAttachV2ParseID(state.Primary.ID)
	if err != nil {
		return nil, err
	}

	resp, err := nics.Get(client, instanceId)
	if err != nil {
		return nil, fmtp.Errorf("Error retrieving ECS nics: %s", err)
	}

	attachment := ecs.GetNicOfEcs(resp, portId)
	if attachment == nil {
		return nil, fmtp.Errorf("Error retrieving ECS nics: UNFOUND")
	}

	return attachment, nil
}

func TestAccResourceEcsNic_basic(t *testing.T) {
	var instance nics.CreateOps
	resourceName := "huaweicloud_compute_interface_attach.test"
	name := acceptance.RandomAccResourceNameWithDash()

	rc := acceptance.InitResourceCheck(
		resourceName,
		&instance,
		getEcsNictResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeV2InterfaceAttach_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fixed_ip", "192.168.0.199"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_id"},
			},
		},
	})
}

const testAccCompute_data = `
data "huaweicloud_availability_zones" "test" {}

data "huaweicloud_compute_flavors" "test" {
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  performance_type  = "normal"
  cpu_core_count    = 2
  memory_size       = 4
}

data "huaweicloud_vpc_subnet" "test" {
  name = "subnet-default"
}

data "huaweicloud_images_image" "test" {
  name        = "Ubuntu 18.04 server 64bit"
  most_recent = true
}

data "huaweicloud_networking_secgroup" "test" {
  name = "default"
}
`

func testAccComputeV2InterfaceAttach_basic(rName string) string {
	return fmt.Sprintf(`
%s

resource "huaweicloud_compute_instance" "instance_1" {
  name               = "%s"
  image_id           = data.huaweicloud_images_image.test.id
  flavor_id          = data.huaweicloud_compute_flavors.test.ids[0]
  security_group_ids = [data.huaweicloud_networking_secgroup.test.id]
  availability_zone  = data.huaweicloud_availability_zones.test.names[0]
  network {
    uuid = data.huaweicloud_vpc_subnet.test.id
  }
}

resource "huaweicloud_compute_interface_attach" "ai_1" {
  instance_id = huaweicloud_compute_instance.instance_1.id
  subnet_id  = data.huaweicloud_vpc_subnet.test.id
  fixed_ip    = "192.168.0.199"
  security_group_id = data.huaweicloud_networking_secgroup.test.id
}
`, testAccCompute_data, rName)
}
