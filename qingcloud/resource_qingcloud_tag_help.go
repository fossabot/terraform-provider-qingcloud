package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func modifyTagAttributes(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).tag
	input := new(qc.ModifyTagAttributesInput)
	input.Tag = qc.String(d.Id())
	attributeUpdate := false
	descriptionUpdate := false
	input.TagName, attributeUpdate = getNamePointer(d)
	input.Description, descriptionUpdate = getDescriptionPointer(d)
	if d.HasChange("color") {
		input.Color = qc.String(d.Get("color").(string))
		attributeUpdate = true
	}
	if attributeUpdate || descriptionUpdate {
		var output *qc.ModifyTagAttributesOutput
		var err error
		simpleRetry(func() error {
			output, err = clt.ModifyTagAttributes(input)
			return isServerBusy(err)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceSetTag(d *schema.ResourceData, tags []*qc.Tag) {
	tagIDs := make([]string, 0, len(tags))
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagIDs = append(tagIDs, qc.StringValue(tag.TagID))
		tagNames = append(tagNames, qc.StringValue(tag.TagName))
	}
	d.Set(resourceTagIds, tagIDs)
	d.Set(resourceTagNames, tagNames)
}

func resourceUpdateTag(d *schema.ResourceData, meta interface{}, resourceType string) error {
	if !d.HasChange(resourceTagIds) {
		return nil
	}
	clt := meta.(*QingCloudClient).tag
	oldV, newV := d.GetChange(resourceTagIds)
	var oldTags []string
	var newTags []string
	for _, v := range oldV.(*schema.Set).List() {
		oldTags = append(oldTags, v.(string))
	}
	for _, v := range newV.(*schema.Set).List() {
		newTags = append(newTags, v.(string))
	}
	attachTags, detachTags := stringSliceDiff(newTags, oldTags)

	if len(detachTags) > 0 {
		input := new(qc.DetachTagsInput)
		for _, tag := range detachTags {
			rtp := qc.ResourceTagPair{
				TagID:        qc.String(tag),
				ResourceID:   qc.String(d.Id()),
				ResourceType: qc.String(resourceType),
			}
			input.ResourceTagPairs = append(input.ResourceTagPairs, &rtp)
		}
		var output *qc.DetachTagsOutput
		var err error
		simpleRetry(func() error {
			output, err = clt.DetachTags(input)
			return isServerBusy(err)
		})
		if err != nil {
			return err
		}
	}
	if len(attachTags) > 0 {
		input := new(qc.AttachTagsInput)
		for _, tag := range attachTags {
			rtp := qc.ResourceTagPair{
				TagID:        qc.String(tag),
				ResourceID:   qc.String(d.Id()),
				ResourceType: qc.String(resourceType),
			}
			input.ResourceTagPairs = append(input.ResourceTagPairs, &rtp)
		}
		var output *qc.AttachTagsOutput
		var err error
		simpleRetry(func() error {
			output, err = clt.AttachTags(input)
			return isServerBusy(err)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func tagIdsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
	}
}

func tagNamesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
	}
}
