package huaweicloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/compute/v2/extensions/attachinterfaces"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/nics"
	"github.com/chnsz/golangsdk/openstack/networking/v2/ports"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

func ResourceComputeInterfaceAttachV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeInterfaceAttachV2Create,
		Read:          resourceComputeInterfaceAttachV2Read,
		Update:        resourceComputeInterfaceAttachV2Update,
		Delete:        resourceComputeInterfaceAttachV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"network_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"port_id", "network_id", "subnet_id"},
				Deprecated:   "Please use subnet_id instead.",
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"port_id", "network_id", "subnet_id"},
			},

			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mac": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"attachable_quantity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceComputeInterfaceAttachV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)
	computeClient, err := config.ComputeV2Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating compute client: %s", err)
	}

	instanceId := d.Get("instance_id").(string)

	var subnetId string
	if v, ok := d.GetOk("network_id"); ok {
		subnetId = v.(string)
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		subnetId = v.(string)
	}

	var securityGroupIds []nics.IdInfo
	if v, ok := d.GetOk("security_group_id"); ok {
		securityGroupIds = append(securityGroupIds, nics.IdInfo{Id: v.(string)})
	}

	opts := nics.CreateOps{
		Nics: []nics.NicReq{
			{
				SubnetId:       subnetId,
				IpAddress:      d.Get("fixed_ip").(string),
				SecurityGroups: securityGroupIds,
			},
		},
	}

	createRst := nics.Create(computeClient, instanceId, opts)
	if createRst.Err != nil {
		return fmtp.DiagErrorf("Error creating huaweicloud_compute_interface_attach: %s", createRst.Err)
	}

	portID := "attachment.PortID"
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ATTACHING"},
		Target:     []string{"ATTACHED"},
		Refresh:    computeInterfaceAttachV2AttachFunc(computeClient, instanceId, portID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return diag.Errorf("Error creating huaweicloud_compute_interface_attach %s: %s", instanceId, err)
	}

	// Use the instance ID and port ID as the resource ID.
	id := fmt.Sprintf("%s/%s", instanceId, portID)

	//	logp.Printf("[DEBUG] Created huaweicloud_compute_interface_attach %s: %#v", id, attachment)

	d.SetId(id)

	if sourceDestCheck := d.Get("source_dest_check").(bool); !sourceDestCheck {
		nicClient, err := config.NetworkingV2Client(GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating HuaweiCloud networking client: %s", err)
		}
		if err := disableSourceDestCheck(nicClient, portID); err != nil {
			return diag.Errorf("Error disable source dest check on port(%s) of instance(%s) failed: %s", portID, d.Id(), err)
		}
	}

	return nil
}

func resourceComputeInterfaceAttachV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*config.Config)
	computeClient, err := config.ComputeV2Client(GetRegion(d, config))
	if err != nil {
		return fmtp.Errorf("Error creating HuaweiCloud compute client: %s", err)
	}
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmtp.Errorf("Error creating HuaweiCloud networking client: %s", err)
	}

	instanceId, portId, err := computeInterfaceAttachV2ParseID(d.Id())
	if err != nil {
		return err
	}

	attachment, err := attachinterfaces.Get(computeClient, instanceId, portId).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving huaweicloud_compute_interface_attach")
	}

	logp.Printf("[DEBUG] Retrieved huaweicloud_compute_interface_attach %s: %#v", d.Id(), attachment)

	d.Set("instance_id", instanceId)
	d.Set("port_id", attachment.PortID)
	d.Set("network_id", attachment.NetID)
	d.Set("region", GetRegion(d, config))

	if len(attachment.FixedIPs) > 0 {
		firstAddress := attachment.FixedIPs[0].IPAddress
		d.Set("fixed_ip", firstAddress)
	}

	if port, err := ports.Get(networkingClient, attachment.PortID).Extract(); err == nil {
		d.Set("source_dest_check", len(port.AllowedAddressPairs) == 0)
		d.Set("mac", port.MACAddress)
	}

	return nil
}

func resourceComputeInterfaceAttachV2Update(d *schema.ResourceData, meta interface{}) error {
	var err error

	config := meta.(*config.Config)
	nicClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmtp.Errorf("Error creating HuaweiCloud networking client: %s", err)
	}

	portID := d.Get("port_id").(string)
	if sourceDestCheck := d.Get("source_dest_check").(bool); !sourceDestCheck {
		err = disableSourceDestCheck(nicClient, portID)
	} else {
		err = enableSourceDestCheck(nicClient, portID)
	}

	if err != nil {
		return fmtp.Errorf("Error updating source_dest_check on port(%s) of instance(%s) failed: %s", portID, d.Id(), err)
	}

	return resourceComputeInterfaceAttachV2Read(d, meta)
}

func resourceComputeInterfaceAttachV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*config.Config)
	computeClient, err := config.ComputeV2Client(GetRegion(d, config))
	if err != nil {
		return fmtp.Errorf("Error creating HuaweiCloud compute client: %s", err)
	}

	instanceId, portId, err := computeInterfaceAttachV2ParseID(d.Id())
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{""},
		Target:     []string{"DETACHED"},
		Refresh:    computeInterfaceAttachV2DetachFunc(computeClient, instanceId, portId),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmtp.Errorf("Error detaching huaweicloud_compute_interface_attach %s: %s", d.Id(), err)
	}

	return nil
}

func computeInterfaceAttachV2AttachFunc(
	computeClient *golangsdk.ServiceClient, instanceId, attachmentId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		va, err := attachinterfaces.Get(computeClient, instanceId, attachmentId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return va, "ATTACHING", nil
			}
			return va, "", err
		}

		return va, "ATTACHED", nil
	}
}

func waitingNotebookForRunning(ctx context.Context, client *golangsdk.ServiceClient, instanceId string, createOpts nics.NicReq, timeout time.Duration) error {
	createStateConf := &resource.StateChangeConf{
		Pending: []string{"ATTACHING"},
		Target:  []string{"ATTACHED"},
		Refresh: func() (interface{}, string, error) {
			resp, err := nics.Get(client, instanceId)
			if err != nil {
				return nil, "failed", err
			}

			return resp, "", err
		},
		Timeout:      timeout,
		PollInterval: 10 * timeout,
		Delay:        10 * time.Second,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for ModelArts notebook (%s) to be created: %s", instanceId, err)
	}
	return nil
}
