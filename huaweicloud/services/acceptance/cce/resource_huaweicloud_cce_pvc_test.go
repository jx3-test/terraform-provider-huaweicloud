package cce

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/cce"

	"github.com/chnsz/golangsdk/openstack/cce/v1/persistentvolumeclaims"
)

func getPvcResourceFunc(conf *config.Config, state *terraform.ResourceState) (interface{}, error) {
	c, err := conf.CceV1Client(acceptance.HW_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating HuaweiCloud CCE v1 client: %s", err)
	}
	resp, err := cce.GetCcePvcInfoById(c, state.Primary.Attributes["cluster_id"],
		state.Primary.Attributes["namespace"], state.Primary.ID)
	if resp == nil && err == nil {
		return resp, fmt.Errorf("Unable to find the persistent volume claim (%s)", state.Primary.ID)
	}
	return resp, err
}

func TestAccCcePersistentVolumeClaimsV1_basic(t *testing.T) {
	var pvc persistentvolumeclaims.PersistentVolumeClaim
	resourceName := "huaweicloud_cce_pvc.test"
	randName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	rc := acceptance.InitResourceCheck(
		resourceName,
		&pvc,
		getPvcResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
		},
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCcePersistentVolumeClaimsV1_basic(randName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					acceptance.TestCheckResourceAttrWithVariable(resourceName, "cluster_id",
						"${huaweicloud_cce_cluster.test.id}"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "default"),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "storage_class_name", "csi-disk"),
				),
			},
		},
	})
}

func TestAccCcePersistentVolumeClaimsV1_obs(t *testing.T) {
	var pvc persistentvolumeclaims.PersistentVolumeClaim
	resourceName := "huaweicloud_cce_pvc.test"
	randName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	rc := acceptance.InitResourceCheck(
		resourceName,
		&pvc,
		getPvcResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
		},
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCcePersistentVolumeClaimsV1_obs(randName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					acceptance.TestCheckResourceAttrWithVariable(resourceName, "cluster_id",
						"${huaweicloud_cce_cluster.test.id}"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "default"),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "storage_class_name", "csi-obs"),
				),
			},
		},
	})
}

func TestAccCcePersistentVolumeClaimsV1_sfs(t *testing.T) {
	var pvc persistentvolumeclaims.PersistentVolumeClaim
	resourceName := "huaweicloud_cce_pvc.test"
	randName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	rc := acceptance.InitResourceCheck(
		resourceName,
		&pvc,
		getPvcResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
		},
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCcePersistentVolumeClaimsV1_sfs(randName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					acceptance.TestCheckResourceAttrWithVariable(resourceName, "cluster_id",
						"${huaweicloud_cce_cluster.test.id}"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "default"),
					resource.TestCheckResourceAttr(resourceName, "name", randName),
					resource.TestCheckResourceAttr(resourceName, "storage_class_name", "csi-nas"),
				),
			},
		},
	})
}

func testAccCceCluster_config(rName string) string {
	return fmt.Sprintf(`
data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_vpc" "test" {
  name = "%s"
  cidr = "192.168.0.0/20"
}

resource "huaweicloud_vpc_subnet" "test" {
  name       = "%s"
  cidr       = "192.168.0.0/24"
  gateway_ip = "192.168.0.1"
  vpc_id     = huaweicloud_vpc.test.id
}

resource "huaweicloud_compute_keypair" "test" {
  name = "%s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDAjpC1hwiOCCmKEWxJ4qzTTsJbKzndLo1BCz5PcwtUnflmU+gHJtWMZKpuEGVi29h0A/+ydKek1O18k10Ff+4tyFjiHDQAT9+OfgWf7+b1yK+qDip3X1C0UPMbwHlTfSGWLGZquwhvEFx9k3h/M+VtMvwR1lJ9LUyTAImnNjWG7TAIPmui30HvM2UiFEmqkr4ijq45MyX2+fLIePLRIFuu1p4whjHAQYufqyno3BS48icQb4p6iVEZPo4AE2o9oIyQvj2mx4dk5Y8CgSETOZTYDOR3rU2fZTRDRgPJDH9FWvQjF5tA0p3d9CoWWd2s6GKKbfoUIi8R/Db1BSPJwkqB jrp-hp-pc"
}

resource "huaweicloud_cce_cluster" "test" {
  name                   = "%s"
  flavor_id              = "cce.s1.small"
  vpc_id                 = huaweicloud_vpc.test.id
  subnet_id              = huaweicloud_vpc_subnet.test.id
  container_network_type = "overlay_l2"
}

resource "huaweicloud_cce_node" "test" {
  cluster_id        = huaweicloud_cce_cluster.test.id
  name              = "%s"
  flavor_id         = "s6.large.2"
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  key_pair          = huaweicloud_compute_keypair.test.name

  root_volume {
    size       = 40
    volumetype = "SSD"
  }
  data_volumes {
    size       = 100
    volumetype = "SSD"
  }
}`, rName, rName, rName, rName, rName)
}

func testAccCcePersistentVolumeClaimsV1_basic(rName string) string {
	return fmt.Sprintf(`
%s

resource "huaweicloud_cce_pvc" "test" {
  cluster_id  = huaweicloud_cce_cluster.test.id
  namespace   = "default"
  name        = "%s"
  annotations = {
    "everest.io/disk-volume-type" = "SSD"
  }
  storage_class_name = "csi-disk"
  access_modes = ["ReadWriteOnce"]
  storage = "10Gi"
}
`, testAccCceCluster_config(rName), rName)
}

func testAccCcePersistentVolumeClaimsV1_obs(rName string) string {
	return fmt.Sprintf(`
%s

resource "huaweicloud_cce_pvc" "test" {
  cluster_id  = huaweicloud_cce_cluster.test.id
  namespace   = "default"
  name        = "%s"
  annotations = {
    "everest.io/obs-volume-type" = "STANDARD"
    "csi.storage.k8s.io/fstype" =  "obsfs"
  }
  storage_class_name = "csi-obs"
  access_modes = ["ReadWriteMany"]
  storage = "1Gi"
}
`, testAccCceCluster_config(rName), rName)
}

func testAccCcePersistentVolumeClaimsV1_sfs(rName string) string {
	return fmt.Sprintf(`
%s

resource "huaweicloud_cce_pvc" "test" {
  cluster_id  = huaweicloud_cce_cluster.test.id
  namespace   = "default"
  name        = "%s"
  storage_class_name = "csi-nas"
  access_modes = ["ReadWriteMany"]
  storage = "10Gi"
}
`, testAccCceCluster_config(rName), rName)
}
