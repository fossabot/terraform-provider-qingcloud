package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func resourceQingcloudInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudInstanceCreate,
		Read:   resourceQingcloudInstanceRead,
		Update: resourceQingcloudInstanceUpdate,
		Delete: resourceQingcloudInstanceDelete,
		Schema: map[string]*schema.Schema{
			resourceName: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of instance",
			},
			resourceDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of instance",
			},
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpu": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: withinArrayInt(1, 2, 4, 8, 16),
				Default:      1,
			},
			"memory": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: withinArrayInt(1024, 2048, 4096, 6144, 8192, 12288, 16384, 24576, 32768),
				Default:      1024,
			},
			"instance_class": &schema.Schema{
				Type:         schema.TypeInt,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: withinArrayInt(0, 1),
				Default:      0,
				Description:  "Type of instance , 0 - Performance type , 1 - Ultra high performance type",
			},
			"managed_vxnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "vxnet-0",
			},
			"private_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"keypair_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"eip_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"volume_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"public_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceTagIds:   tagIdsSchema(),
			resourceTagNames: tagNamesSchema(),
		},
	}
}

func resourceQingcloudInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance
	input := new(qc.RunInstancesInput)
	input.Count = qc.Int(1)
	input.InstanceName, _ = getNamePointer(d)
	input.ImageID = qc.String(d.Get("image_id").(string))
	input.CPU = qc.Int(d.Get("cpu").(int))
	input.Memory = qc.Int(d.Get("memory").(int))
	input.InstanceClass = qc.Int(d.Get("instance_class").(int))
	input.SecurityGroup = getSetStringPointer(d, "security_group_id")
	input.LoginMode = qc.String("keypair")
	kps := d.Get("keypair_ids").(*schema.Set).List()
	if len(kps) > 0 {
		kp := kps[0].(string)
		input.LoginKeyPair = qc.String(kp)
	}
	var output *qc.RunInstancesOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.RunInstances(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	d.SetId(qc.StringValue(output.Instances[0]))
	if _, err := InstanceTransitionStateRefresh(clt, d.Id()); err != nil {
		return err
	}
	return resourceQingcloudInstanceUpdate(d, meta)
}

func resourceQingcloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).instance
	input := new(qc.DescribeInstancesInput)
	input.Instances = []*string{qc.String(d.Id())}
	input.Verbose = qc.Int(1)
	var output *qc.DescribeInstancesOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.DescribeInstances(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if isInstanceDeleted(output.InstanceSet) {
		d.SetId("")
		return nil
	}
	instance := output.InstanceSet[0]
	d.Set(resourceName, qc.StringValue(instance.InstanceName))
	d.Set(resourceDescription, qc.StringValue(instance.Description))
	d.Set("image_id", qc.StringValue(instance.Image.ImageID))
	d.Set("cpu", qc.IntValue(instance.VCPUsCurrent))
	d.Set("memory", qc.IntValue(instance.MemoryCurrent))
	d.Set("instance_class", qc.IntValue(instance.InstanceClass))
	//set managed vxnet
	for _, vxnet := range instance.VxNets {
		if qc.IntValue(vxnet.VxNetType) != 0 {
			if qc.IntValue(vxnet.VxNetType) == 1 {
				d.Set("managed_vxnet_id", qc.StringValue(vxnet.VxNetID))
				d.Set("private_ip", qc.StringValue(vxnet.PrivateIP))
			} else {
				d.Set("managed_vxnet_id", "vxnet-0")
				d.Set("private_ip", qc.StringValue(vxnet.PrivateIP))
			}
		}
	}
	if instance.EIP != nil {
		d.Set("eip_id", qc.StringValue(instance.EIP.EIPID))
		d.Set("public_ip", qc.StringValue(instance.EIP.EIPAddr))
	}
	if instance.SecurityGroup != nil {
		d.Set("security_group_id", qc.StringValue(instance.SecurityGroup.SecurityGroupID))
	}
	if instance.KeyPairIDs != nil {
		keypairIDs := make([]string, 0, len(instance.KeyPairIDs))
		for _, kp := range instance.KeyPairIDs {
			keypairIDs = append(keypairIDs, qc.StringValue(kp))
		}
		d.Set("keypair_ids", keypairIDs)
	}
	if instance.Volumes != nil {
		volumeIDs := make([]string, 0, len(instance.Volumes))
		for _, volume := range instance.Volumes {
			volumeIDs = append(volumeIDs, qc.StringValue(volume.VolumeID))
		}
		d.Set("volume_ids", volumeIDs)
	}
	resourceSetTag(d, instance.Tags)
	return nil
}

func resourceQingcloudInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
	if err := waitInstanceLease(d, meta); err != nil {
		return err
	}
	if err := modifyInstanceAttributes(d, meta); err != nil {
		return err
	}
	d.SetPartial("name")
	d.SetPartial("description")
	// change vxnet
	if err := instanceUpdateChangeManagedVxNet(d, meta); err != nil {
		return err
	}
	d.SetPartial("managed_vxnet_id")
	d.SetPartial("private_ip")
	// change security_group
	if err := instanceUpdateChangeSecurityGroup(d, meta); err != nil {
		return err
	}
	d.SetPartial("security_group_id")
	// change eip
	if err := instanceUpdateChangeEip(d, meta); err != nil {
		return err
	}
	d.SetPartial("eip_id")
	// change keypairs
	if err := instanceUpdateChangeKeyPairs(d, meta); err != nil {
		return err
	}
	d.SetPartial("keypair_ids")
	// change volumes
	if err := updateInstanceVolume(d, meta); err != nil {
		return err
	}
	d.SetPartial("volume_ids")
	// resize instance
	if err := instanceUpdateResize(d, meta); err != nil {
		return err
	}
	d.SetPartial("cpu")
	d.SetPartial("memory")
	if err := resourceUpdateTag(d, meta, qingcloudResourceTypeInstance); err != nil {
		return err
	}
	d.Partial(false)
	return resourceQingcloudInstanceRead(d, meta)
}

func resourceQingcloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	if err := waitInstanceLease(d, meta); err != nil {
		return err
	}
	clt := meta.(*QingCloudClient).instance
	input := new(qc.TerminateInstancesInput)
	input.Instances = []*string{qc.String(d.Id())}
	var output *qc.TerminateInstancesOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.TerminateInstances(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if _, err := InstanceTransitionStateRefresh(clt, d.Id()); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
