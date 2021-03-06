package qingcloud

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func TestAccQingcloudInstance_basic(t *testing.T) {
	var instance qc.DescribeInstancesOutput
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "qingcloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("qingcloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "image_id", "centos7x64d"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "managed_vxnet_id", "vxnet-0"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "cpu", "1"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "instance_class", "0"),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("qingcloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", resourceName, "instance"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", resourceDescription, "instance"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "image_id", "centos7x64d"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "managed_vxnet_id", "vxnet-0"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "cpu", "2"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "memory", "2048"),
					resource.TestCheckResourceAttr(
						"qingcloud_instance.foo", "instance_class", "0"),
				),
			},
		},
	})
}

func TestAccQingcloudInstance_tag(t *testing.T) {
	var instance qc.DescribeInstancesOutput
	instanceTag1Name := os.Getenv("TRAVIS_BUILD_ID") + "-" + os.Getenv("TRAVIS_JOB_NUMBER") + "-instance-tag1"
	instanceTag2Name := os.Getenv("TRAVIS_BUILD_ID") + "-" + os.Getenv("TRAVIS_JOB_NUMBER") + "-instance-tag2"
	testTagNameValue := func(names ...string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			tags := instance.InstanceSet[0].Tags
			same_count := 0
			for _, tag := range tags {
				for _, name := range names {
					if qc.StringValue(tag.TagName) == name {
						same_count++
					}
					if same_count == len(instance.InstanceSet[0].Tags) {
						return nil
					}
				}
			}
			return fmt.Errorf("tag name error %#v", names)
		}
	}
	testTagDetach := func() resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if len(instance.InstanceSet[0].Tags) != 0 {
				return fmt.Errorf("tag not detach ")
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "qingcloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccInstanceConfigTagTemplate, instanceTag1Name, instanceTag2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testTagNameValue(instanceTag1Name, instanceTag2Name),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccInstanceConfigTagTwoTemplate, instanceTag1Name, instanceTag2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testTagDetach(),
				),
			},
		},
	})

}

func TestAccQingcloudInstance_multiKeypairByCount(t *testing.T) {
	var instance qc.DescribeInstancesOutput
	testCheck := func(kpCount int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(instance.InstanceSet[0].KeyPairIDs) < 0 {
				return fmt.Errorf("no keypairs: %#v", instance.InstanceSet[0].KeyPairIDs)
			}

			if len(instance.InstanceSet[0].KeyPairIDs) != kpCount {
				return fmt.Errorf("keypair count inconformity : %#v", instance.InstanceSet[0].KeyPairIDs)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "qingcloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigKeyPair,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(1),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigKeyPairTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(3),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigKeyPairThree,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(2),
				),
			},
		},
	})

}

func TestAccQingcloudInstance_multiVolumeByCount(t *testing.T) {
	var instance qc.DescribeInstancesOutput
	testCheck := func(volCount int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(instance.InstanceSet[0].Volumes) < 0 {
				return fmt.Errorf("no keypairs: %#v", instance.InstanceSet[0].Volumes)
			}

			if len(instance.InstanceSet[0].Volumes) != volCount {
				return fmt.Errorf("keypair count inconformity : %#v", instance.InstanceSet[0].Volumes)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "qingcloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(1),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigVolumeTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(3),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigVolumeThree,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testCheck(2),
				),
			},
		},
	})

}

func TestAccQingcloudInstance_eip(t *testing.T) {
	var instance qc.DescribeInstancesOutput

	testEIP := func() resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if instance.InstanceSet[0].EIP != nil {
				input := new(qc.DescribeEIPsInput)
				input.EIPs = []*string{instance.InstanceSet[0].EIP.EIPID}
				input.Verbose = qc.Int(1)
				client := testAccProvider.Meta().(*QingCloudClient)
				d, err := client.eip.DescribeEIPs(input)

				if err != nil {
					return err
				}
				if d == nil || len(d.EIPSet) == 0 {
					return fmt.Errorf("EIP not found ")
				}
				if qc.StringValue(d.EIPSet[0].EIPAddr) != qc.StringValue(instance.InstanceSet[0].EIP.EIPAddr) {
					return fmt.Errorf("EIP not match ")
				}
				return nil
			} else {
				return fmt.Errorf("Can not find eip ")
			}
		}
	}
	testDetachEIP := func() resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if instance.InstanceSet[0].EIP != nil && instance.InstanceSet[0].EIP.EIPID != nil && qc.StringValue(instance.InstanceSet[0].EIP.EIPID) != "" {
				return fmt.Errorf("EIP not detach ")
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "qingcloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigEIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testEIP(),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigEIPTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"qingcloud_instance.foo", &instance),
					testDetachEIP(),
				),
			},
		},
	})

}

func testAccCheckInstanceExists(n string, i *qc.DescribeInstancesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Instance ID is set")
		}

		client := testAccProvider.Meta().(*QingCloudClient)
		input := new(qc.DescribeInstancesInput)
		input.Verbose = qc.Int(1)
		input.Instances = []*string{qc.String(rs.Primary.ID)}
		d, err := client.instance.DescribeInstances(input)

		log.Printf("[WARN] instance id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		if d == nil || len(d.InstanceSet) == 0 {
			return fmt.Errorf("Instance not found ")
		}

		*i = *d
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	return testAccCheckInstanceWithProvider(s, testAccProvider)
}

func testAccCheckInstanceWithProvider(s *terraform.State, provider *schema.Provider) error {
	client := provider.Meta().(*QingCloudClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "qingcloud_instance" {
			continue
		}
		// Try to find the resource
		input := new(qc.DescribeInstancesInput)
		input.Instances = []*string{qc.String(rs.Primary.ID)}
		output, err := client.instance.DescribeInstances(input)
		if err == nil {
			if !isInstanceDeleted(output.InstanceSet) {
				return fmt.Errorf("Found  Instance: %s", rs.Primary.ID)
			}
		}
	}
	return nil
}

const testAccInstanceConfig = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`
const testAccInstanceConfigTwo = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
	cpu = 2
    memory = 2048
	name = "instance"
	description = "instance"
}
`
const testAccInstanceConfigTagTemplate = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
	tag_ids = ["${qingcloud_tag.test.id}",
				"${qingcloud_tag.test2.id}"]
}
resource "qingcloud_tag" "test"{
	name="%v"
}
resource "qingcloud_tag" "test2"{
	name="%v"
}
`
const testAccInstanceConfigTagTwoTemplate = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
resource "qingcloud_tag" "test"{
	name="%v"
}
resource "qingcloud_tag" "test2"{
	name="%v"
}
`
const testAccInstanceConfigKeyPair = `
resource "qingcloud_keypair" "foo1"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo2"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo3"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo1.id}"]
}
`

const testAccInstanceConfigKeyPairTwo = `
resource "qingcloud_keypair" "foo1"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo2"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo3"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo1.id}","${qingcloud_keypair.foo2.id}","${qingcloud_keypair.foo3.id}"]
}
`
const testAccInstanceConfigKeyPairThree = `
resource "qingcloud_keypair" "foo1"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo2"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_keypair" "foo3"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo1.id}","${qingcloud_keypair.foo2.id}"]
}
`

const testAccInstanceConfigVolume = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_volume" "foo1"{
	size = 10
}
resource "qingcloud_volume" "foo2"{
	size = 10
}
resource "qingcloud_volume" "foo3"{
	size = 10
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	volume_ids = ["${qingcloud_volume.foo1.id}"]
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`

const testAccInstanceConfigVolumeTwo = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_volume" "foo1"{
	size = 10
}
resource "qingcloud_volume" "foo2"{
	size = 10
}
resource "qingcloud_volume" "foo3"{
	size = 10
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	volume_ids = ["${qingcloud_volume.foo1.id}","${qingcloud_volume.foo2.id}","${qingcloud_volume.foo3.id}"]
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`
const testAccInstanceConfigVolumeThree = `
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_volume" "foo1"{
	size = 10
}
resource "qingcloud_volume" "foo2"{
	size = 10
}
resource "qingcloud_volume" "foo3"{
	size = 10
}
resource "qingcloud_instance" "foo" {
	image_id = "centos7x64d"
	volume_ids = ["${qingcloud_volume.foo1.id}","${qingcloud_volume.foo2.id}"]
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`

const testAccInstanceConfigEIP = `
resource "qingcloud_security_group" "foo" {
    name = "first_sg"
}
resource "qingcloud_eip" "foo" {
    bandwidth = 2
}
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	security_group_id = "${qingcloud_security_group.foo.id}"
	image_id = "centos7x64d"
	eip_id = "${qingcloud_eip.foo.id}"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`
const testAccInstanceConfigEIPTwo = `
resource "qingcloud_security_group" "foo" {
    name = "first_sg"
}
resource "qingcloud_eip" "foo" {
    bandwidth = 2
}
resource "qingcloud_keypair" "foo"{
	public_key = "    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCyLSPqVIdXGH0QlGeWcPwa1fjTRKl6WtMiaSsP8/GnwjakDSKILUCoNe1yIpiK8F0/gmL71xaDQyfl7k6aE+gn6lSLUjpDmucAF1luGg6l7CIN+6hCqY3YqlAI05Tqwu0PdLAwCbGwdHcaWfECcbROJk5D0zpCTHmissrrAxdOv72g9Ple8KJ6C7F1tz6wmG0zUeineguGjW/PvfZiBDWZ/CyXGPeMDJxv3lrIiLa/ShgnQOxFTdHJPCw+F0/XlSzlIzP3gfni1vXxJWvYjdE9ULo7Z1DLWgZ73FCbeAvX/0e9C9jwT21Qa5RUy4pSP8m4WXSJgw2f9IpY1vIJFSZP root@centos1    "
}
resource "qingcloud_instance" "foo" {
	security_group_id = "${qingcloud_security_group.foo.id}"
	image_id = "centos7x64d"
	keypair_ids = ["${qingcloud_keypair.foo.id}"]
}
`
