package qingcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func resourceQingcloudSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudSecurityGroupCreate,
		Read:   resourceQingcloudSecurityGroupRead,
		Update: resourceQingcloudSecurityGroupUpdate,
		Delete: resourceQingcloudSecurityGroupDelete,
		Schema: map[string]*schema.Schema{
			resourceName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of SecurityGroup ",
			},
			resourceDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of SecurityGroup",
			},
			resourceTagIds:   tagIdsSchema(),
			resourceTagNames: tagNamesSchema(),
		},
	}
}

func resourceQingcloudSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.CreateSecurityGroupInput)
	input.SecurityGroupName, _ = getNamePointer(d)
	var output *qc.CreateSecurityGroupOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.CreateSecurityGroup(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	d.SetId(qc.StringValue(output.SecurityGroupID))
	return resourceQingcloudSecurityGroupUpdate(d, meta)
}

func resourceQingcloudSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.DescribeSecurityGroupsInput)
	input.SecurityGroups = []*string{qc.String(d.Id())}
	var output *qc.DescribeSecurityGroupsOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.DescribeSecurityGroups(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if len(output.SecurityGroupSet) == 0 {
		d.SetId("")
		return nil
	}
	sg := output.SecurityGroupSet[0]
	d.Set(resourceName, qc.StringValue(sg.SecurityGroupName))
	d.Set(resourceDescription, qc.StringValue(sg.Description))
	resourceSetTag(d, sg.Tags)
	return nil
}
func resourceQingcloudSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
	if err := modifySecurityGroupAttributes(d, meta); err != nil {
		return err
	}
	d.SetPartial(resourceDescription)
	d.SetPartial(resourceName)
	if err := resourceUpdateTag(d, meta, qingcloudResourceTypeSecurityGroup); err != nil {
		return err
	}
	d.SetPartial(resourceTagIds)
	d.Partial(false)
	return resourceQingcloudSecurityGroupRead(d, meta)
}

func resourceQingcloudSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	describeSecurityGroupInput := new(qc.DescribeSecurityGroupsInput)
	describeSecurityGroupInput.SecurityGroups = []*string{qc.String(d.Id())}
	describeSecurityGroupInput.Verbose = qc.Int(1)
	var describeSecurityGroupOutput *qc.DescribeSecurityGroupsOutput
	var err error
	simpleRetry(func() error {
		describeSecurityGroupOutput, err = clt.DescribeSecurityGroups(describeSecurityGroupInput)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if len(describeSecurityGroupOutput.SecurityGroupSet[0].Resources) > 0 {
		return fmt.Errorf("Error security group %s is using, can't delete", d.Id())
	}
	input := new(qc.DeleteSecurityGroupsInput)
	input.SecurityGroups = []*string{qc.String(d.Id())}
	var output *qc.DeleteSecurityGroupsOutput
	simpleRetry(func() error {
		output, err = clt.DeleteSecurityGroups(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
