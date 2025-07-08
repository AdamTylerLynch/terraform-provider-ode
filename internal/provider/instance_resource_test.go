// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccODEInstance_basic(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	config := testAccProviderConfig() + testAccInstanceResourceConfig_basic()

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and test
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("label"),
							knownvalue.StringExact(os.Getenv("INSTANCE_LABEL")),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("target_uuid"),
							knownvalue.StringExact(os.Getenv("INSTANCE_TARGET_UUID")),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("image_uuid"),
							knownvalue.StringExact(os.Getenv("INSTANCE_IMAGE_UUID")),
						),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("provision_uuid")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("hostname")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("status")),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.label", os.Getenv("INSTANCE_LABEL")),
			resource.TestCheckResourceAttr("ode_instance.test", "general.target_uuid", os.Getenv("INSTANCE_TARGET_UUID")),
			resource.TestCheckResourceAttr("ode_instance.test", "general.image_uuid", os.Getenv("INSTANCE_IMAGE_UUID")),
			resource.TestCheckResourceAttrSet("ode_instance.test", "provision_uuid"),
			resource.TestCheckResourceAttrSet("ode_instance.test", "hostname"),
			resource.TestCheckResourceAttr("ode_instance.test", "status", "completed"),
		)

		// Import test only in full mode
		testCase.Steps = append(testCase.Steps, resource.TestStep{
			ResourceName:      "ode_instance.test",
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateVerifyIgnore: []string{
				"ssh_target_user",
				"ssh_target_password",
				"ssh_target_key_file",
				"ssh_target_passphrase",
			},
		})
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_disappears(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	config := testAccProviderConfig() + testAccInstanceResourceConfig_basic()

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("provision_uuid")),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("ode_instance.test", "provision_uuid"),
			testAccCheckInstanceDisappears("ode_instance.test"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_description(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	configWithoutDesc := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithDesc := testAccProviderConfig() + testAccInstanceResourceConfig_withDescription("Test instance description")
	configUpdatedDesc := testAccProviderConfig() + testAccInstanceResourceConfig_withDescription("Updated description")

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without description
			{
				Config: configWithoutDesc,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add description (requires replace)
			{
				Config: configWithDesc,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("description"),
							knownvalue.StringExact("Test instance description"),
						),
					},
				},
			},
			// Update description (requires replace)
			{
				Config: configUpdatedDesc,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("description"),
							knownvalue.StringExact("Updated description"),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "general.description"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.description", "Test instance description"),
		)
		testCase.Steps[2].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.description", "Updated description"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_sshPublicKey(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" || os.Getenv("INSTANCE_SSH_PUBLIC_KEY") == "" {
		t.Fatal("SSH_PASSWORD and INSTANCE_SSH_PUBLIC_KEY environment variables must be set for test")
	}
	configWithoutKey := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithKey := testAccProviderConfig() + testAccInstanceResourceConfig_withSSHPublicKey()

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without SSH public key
			{
				Config: configWithoutKey,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add SSH public key (requires replace if configured)
			{
				Config: configWithKey,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("ssh_public_key"),
							knownvalue.StringExact(os.Getenv("INSTANCE_SSH_PUBLIC_KEY")),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "general.ssh_public_key"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.ssh_public_key", os.Getenv("INSTANCE_SSH_PUBLIC_KEY")),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_deploymentDirectory(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" || os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY") == "" {
		t.Fatal("SSH_PASSWORD and INSTANCE_DEPLOYMENT_DIRECTORY environment variables must be set for test")
	}
	configWithoutDir := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithDir := testAccProviderConfig() + testAccInstanceResourceConfig_withDeploymentDirectory()

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without deployment directory
			{
				Config: configWithoutDir,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add deployment directory (requires replace)
			{
				Config: configWithDir,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("deployment_directory"),
							knownvalue.StringExact(os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY")),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "general.deployment_directory"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.deployment_directory", os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY")),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_sysresComponentUUID(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" || os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID") == "" {
		t.Fatal("SSH_PASSWORD and INSTANCE_SYSRES_COMPONENT_UUID environment variables must be set for test")
	}
	configWithoutSysres := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithSysres := testAccProviderConfig() + testAccInstanceResourceConfig_withSysresComponentUUID()

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without sysres component UUID
			{
				Config: configWithoutSysres,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add sysres component UUID (requires replace)
			{
				Config: configWithSysres,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact(os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID")),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "general.sysres_component_uuid"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "general.sysres_component_uuid", os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID")),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_emulator(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	configWithoutEmulator := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithEmulator := testAccProviderConfig() + testAccInstanceResourceConfig_withEmulator(2, 4294967296, 0)
	configWithEmulatorUpdated := testAccProviderConfig() + testAccInstanceResourceConfig_withEmulator(4, 8589934592, 2)

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without emulator block (optional)
			{
				Config: configWithoutEmulator,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add emulator (requires replace)
			{
				Config: configWithEmulator,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"),
							knownvalue.Int64Exact(2),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"),
							knownvalue.Int64Exact(4294967296),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
							knownvalue.Int64Exact(0),
						),
					},
				},
			},
			// Update emulator values (requires replace)
			{
				Config: configWithEmulatorUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"),
							knownvalue.Int64Exact(4),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"),
							knownvalue.Int64Exact(8589934592),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
							knownvalue.Int64Exact(2),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "emulator.cp"),
			resource.TestCheckNoResourceAttr("ode_instance.test", "emulator.ram"),
			resource.TestCheckNoResourceAttr("ode_instance.test", "emulator.ziip"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.cp", "2"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ram", "4294967296"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ziip", "0"),
		)
		testCase.Steps[2].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.cp", "4"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ram", "8589934592"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ziip", "2"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_emulatorZiip(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	// Test ziip defaults to 0 when not specified
	configWithoutZiip := testAccProviderConfig() + testAccInstanceResourceConfig_withEmulatorNoZiip(2, 4294967296)
	configWithZiip := testAccProviderConfig() + testAccInstanceResourceConfig_withEmulator(2, 4294967296, 1)

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without ziip (should default to 0)
			{
				Config: configWithoutZiip,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"),
							knownvalue.Int64Exact(2),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"),
							knownvalue.Int64Exact(4294967296),
						),
					},
				},
			},
			// Update to set ziip (requires replace)
			{
				Config: configWithZiip,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
							knownvalue.Int64Exact(1),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.cp", "2"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ram", "4294967296"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ziip", "0"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.cp", "2"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ram", "4294967296"),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ziip", "1"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_zosCreds(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" || os.Getenv("ZOS_USERNAME") == "" || os.Getenv("ZOS_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD, ZOS_USERNAME, and ZOS_PASSWORD environment variables must be set for test")
	}
	configWithoutZosCreds := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithZosCreds := testAccProviderConfig() + testAccInstanceResourceConfig_withZosCreds(os.Getenv("ZOS_USERNAME"), os.Getenv("ZOS_PASSWORD"))
	configWithZosCredsUpdated := testAccProviderConfig() + testAccInstanceResourceConfig_withZosCreds("updateduser", "updatedpass")

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without zos_creds (optional)
			{
				Config: configWithoutZosCreds,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add zos_creds (requires replace)
			{
				Config: configWithZosCreds,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("zos_creds").AtMapKey("username"),
							knownvalue.StringExact(os.Getenv("ZOS_USERNAME")),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("zos_creds").AtMapKey("password"),
							knownvalue.StringExact(os.Getenv("ZOS_PASSWORD")),
						),
					},
				},
			},
			// Update zos_creds values (requires replace)
			{
				Config: configWithZosCredsUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("zos_creds").AtMapKey("username"),
							knownvalue.StringExact("updateduser"),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("zos_creds").AtMapKey("password"),
							knownvalue.StringExact("updatedpass"),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "zos_creds.username"),
			resource.TestCheckNoResourceAttr("ode_instance.test", "zos_creds.password"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "zos_creds.username", os.Getenv("ZOS_USERNAME")),
			resource.TestCheckResourceAttr("ode_instance.test", "zos_creds.password", os.Getenv("ZOS_PASSWORD")),
		)
		testCase.Steps[2].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "zos_creds.username", "updateduser"),
			resource.TestCheckResourceAttr("ode_instance.test", "zos_creds.password", "updatedpass"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_ipl(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" || os.Getenv("IPL_DEVICE_ADDRESS") == "" || os.Getenv("IPL_IODF_ADDRESS") == "" || os.Getenv("IPL_LOAD_SUFFIX") == "" {
		t.Fatal("SSH_PASSWORD, IPL_DEVICE_ADDRESS, IPL_IODF_ADDRESS, and IPL_LOAD_SUFFIX environment variables must be set for test")
	}
	configWithoutIPL := testAccProviderConfig() + testAccInstanceResourceConfig_basic()
	configWithIPL := testAccProviderConfig() + testAccInstanceResourceConfig_withIPL(
		os.Getenv("IPL_DEVICE_ADDRESS"),
		os.Getenv("IPL_IODF_ADDRESS"),
		os.Getenv("IPL_LOAD_SUFFIX"),
	)

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without IPL (optional)
			{
				Config: configWithoutIPL,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Update to add IPL (no replace modifier on ipl block)
			{
				Config: configWithIPL,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("ipl").AtMapKey("device_address"),
							knownvalue.StringExact(os.Getenv("IPL_DEVICE_ADDRESS")),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("ipl").AtMapKey("iodf_address"),
							knownvalue.StringExact(os.Getenv("IPL_IODF_ADDRESS")),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("ipl").AtMapKey("load_suffix"),
							knownvalue.StringExact(os.Getenv("IPL_LOAD_SUFFIX")),
						),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckNoResourceAttr("ode_instance.test", "ipl.device_address"),
			resource.TestCheckNoResourceAttr("ode_instance.test", "ipl.iodf_address"),
			resource.TestCheckNoResourceAttr("ode_instance.test", "ipl.load_suffix"),
		)
		testCase.Steps[1].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "ipl.device_address", os.Getenv("IPL_DEVICE_ADDRESS")),
			resource.TestCheckResourceAttr("ode_instance.test", "ipl.iodf_address", os.Getenv("IPL_IODF_ADDRESS")),
			resource.TestCheckResourceAttr("ode_instance.test", "ipl.load_suffix", os.Getenv("IPL_LOAD_SUFFIX")),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_sshKeyAuthentication(t *testing.T) {
	if os.Getenv("SSH_KEY_FILE_PATH") == "" || os.Getenv("INSTANCE_SSH_USER") == "" {
		t.Fatal("SSH_KEY_FILE_PATH and INSTANCE_SSH_USER environment variables must be set for test")
	}
	// Read the key file content
	keyContent, err := os.ReadFile(os.Getenv("SSH_KEY_FILE_PATH"))
	if err != nil {
		t.Fatalf("Failed to read SSH key file: %v", err)
	}

	config := testAccProviderConfig() + testAccInstanceResourceConfig_withSSHKey(string(keyContent))
	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("ssh_target_user"),
							knownvalue.StringExact(os.Getenv("INSTANCE_SSH_USER")),
						),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("provision_uuid")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("status")),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "ssh_target_user", os.Getenv("INSTANCE_SSH_USER")),
			resource.TestCheckResourceAttrSet("ode_instance.test", "provision_uuid"),
			resource.TestCheckResourceAttr("ode_instance.test", "status", "completed"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_sshKeyWithPassphrase(t *testing.T) {
	if os.Getenv("SSH_KEY_FILE_PATH") == "" || os.Getenv("SSH_TARGET_PASSPHRASE") == "" || os.Getenv("INSTANCE_SSH_USER") == "" {
		t.Fatal("SSH_KEY_FILE_PATH, SSH_TARGET_PASSPHRASE, and INSTANCE_SSH_USER environment variables must be set for test")
	}
	// Read the key file content
	keyContent, err := os.ReadFile(os.Getenv("SSH_KEY_FILE_PATH"))
	if err != nil {
		t.Fatalf("Failed to read SSH key file: %v", err)
	}

	config := testAccProviderConfig() + testAccInstanceResourceConfig_withSSHKeyAndPassphrase(string(keyContent), os.Getenv("SSH_TARGET_PASSPHRASE"))
	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("ssh_target_user"),
							knownvalue.StringExact(os.Getenv("INSTANCE_SSH_USER")),
						),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("provision_uuid")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("status")),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ode_instance.test", "ssh_target_user", os.Getenv("INSTANCE_SSH_USER")),
			resource.TestCheckResourceAttrSet("ode_instance.test", "provision_uuid"),
			resource.TestCheckResourceAttr("ode_instance.test", "status", "completed"),
		)
	}

	resource.Test(t, testCase)
}

func TestAccODEInstance_complete(t *testing.T) {
	// This test requires all environment variables to be set
	requiredEnvVars := []string{
		"SSH_PASSWORD",
		"INSTANCE_LABEL",
		"INSTANCE_DESCRIPTION",
		"INSTANCE_TARGET_UUID",
		"INSTANCE_IMAGE_UUID",
		"INSTANCE_SSH_USER",
		"INSTANCE_SSH_PASSWORD",
		"INSTANCE_SSH_PUBLIC_KEY",
		"INSTANCE_SYSRES_COMPONENT_UUID",
		"INSTANCE_DEPLOYMENT_DIRECTORY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("Skipping complete test: %s environment variable not set", envVar)
		}
	}

	params := instanceConfigParams{
		Label:               os.Getenv("INSTANCE_LABEL"),
		Description:         os.Getenv("INSTANCE_DESCRIPTION"),
		TargetUUID:          os.Getenv("INSTANCE_TARGET_UUID"),
		ImageUUID:           os.Getenv("INSTANCE_IMAGE_UUID"),
		SSHUser:             os.Getenv("INSTANCE_SSH_USER"),
		SSHPassword:         os.Getenv("INSTANCE_SSH_PASSWORD"),
		SSHPublicKey:        os.Getenv("INSTANCE_SSH_PUBLIC_KEY"),
		SysresComponentUUID: os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID"),
		DeploymentDirectory: os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY"),
		CP:                  4,
		RAM:                 8589934592, // 8 GiB
		ZIIP:                2,
	}

	config := testAccProviderConfig() + testAccInstanceResourceConfig_full(params)

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and verify all attributes
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),
						// General attributes
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("label"),
							knownvalue.StringExact(params.Label),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("description"),
							knownvalue.StringExact(params.Description),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("target_uuid"),
							knownvalue.StringExact(params.TargetUUID),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("image_uuid"),
							knownvalue.StringExact(params.ImageUUID),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("ssh_public_key"),
							knownvalue.StringExact(params.SSHPublicKey),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact(params.SysresComponentUUID),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("deployment_directory"),
							knownvalue.StringExact(params.DeploymentDirectory),
						),
						// Emulator attributes
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"),
							knownvalue.Int64Exact(int64(params.CP)),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"),
							knownvalue.Int64Exact(int64(params.RAM)),
						),
						plancheck.ExpectKnownValue(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
							knownvalue.Int64Exact(int64(params.ZIIP)),
						),
						// Computed attributes
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("provision_uuid")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("hostname")),
						plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("status")),
					},
				},
			},
		},
	}

	// Comprehensive resource evaluations/checks if not in short mode
	if !testing.Short() {
		testCase.Steps[0].Check = resource.ComposeTestCheckFunc(
			// General attributes
			resource.TestCheckResourceAttr("ode_instance.test", "general.label", params.Label),
			resource.TestCheckResourceAttr("ode_instance.test", "general.description", params.Description),
			resource.TestCheckResourceAttr("ode_instance.test", "general.target_uuid", params.TargetUUID),
			resource.TestCheckResourceAttr("ode_instance.test", "general.image_uuid", params.ImageUUID),
			resource.TestCheckResourceAttr("ode_instance.test", "general.ssh_public_key", params.SSHPublicKey),
			resource.TestCheckResourceAttr("ode_instance.test", "general.sysres_component_uuid", params.SysresComponentUUID),
			resource.TestCheckResourceAttr("ode_instance.test", "general.deployment_directory", params.DeploymentDirectory),

			resource.TestCheckResourceAttr("ode_instance.test", "emulator.cp", strconv.Itoa(params.CP)),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ram", fmt.Sprintf("%d", params.RAM)),
			resource.TestCheckResourceAttr("ode_instance.test", "emulator.ziip", strconv.Itoa(params.ZIIP)),

			// Computed attributes
			resource.TestCheckResourceAttrSet("ode_instance.test", "provision_uuid"),
			resource.TestCheckResourceAttrSet("ode_instance.test", "hostname"),
			resource.TestCheckResourceAttr("ode_instance.test", "status", "completed"),
		)

		// Import test only in full mode
		testCase.Steps = append(testCase.Steps, resource.TestStep{
			ResourceName:      "ode_instance.test",
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateVerifyIgnore: []string{
				"ssh_target_user",
				"ssh_target_password",
				"ssh_target_key_file",
				"ssh_target_passphrase",
			},
		})
	}

	resource.Test(t, testCase)
}

type instanceConfigParams struct {
	Label               string
	Description         string
	TargetUUID          string
	ImageUUID           string
	SSHUser             string
	SSHPassword         string
	SSHPublicKey        string
	SysresComponentUUID string
	DeploymentDirectory string
	CP                  int
	RAM                 int64
	ZIIP                int
}

func testAccInstanceResourceConfig_basic() string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
	)
}

func testAccInstanceResourceConfig_full(p instanceConfigParams) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  emulator = {
    cp   = %d
    ram  = %d
    ziip = %d
  }

  general = {
    label                 = "%s"
    description           = "%s"
    target_uuid           = "%s"
    image_uuid            = "%s"
    ssh_public_key        = "%s"
    sysres_component_uuid = "%s"
    deployment_directory  = "%s"
  }
}
`,
		p.SSHUser,
		p.SSHPassword,
		p.CP,
		p.RAM,
		p.ZIIP,
		p.Label,
		p.Description,
		p.TargetUUID,
		p.ImageUUID,
		p.SSHPublicKey,
		p.SysresComponentUUID,
		p.DeploymentDirectory,
	)
}

func testAccInstanceResourceConfig_withDescription(description string) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    description = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		description,
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
	)
}

func testAccInstanceResourceConfig_withSSHPublicKey() string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label          = "%s"
    target_uuid    = "%s"
    image_uuid     = "%s"
    ssh_public_key = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		os.Getenv("INSTANCE_SSH_PUBLIC_KEY"),
	)
}

func testAccInstanceResourceConfig_withDeploymentDirectory() string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label                = "%s"
    target_uuid          = "%s"
    image_uuid           = "%s"
    deployment_directory = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY"),
	)
}

func testAccInstanceResourceConfig_withSysresComponentUUID() string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label                 = "%s"
    target_uuid           = "%s"
    image_uuid            = "%s"
    sysres_component_uuid = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID"),
	)
}

func testAccInstanceResourceConfig_withEmulator(cp int, ram int64, ziip int) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }

  emulator = {
    cp   = %d
    ram  = %d
    ziip = %d
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		cp,
		ram,
		ziip,
	)
}

func testAccInstanceResourceConfig_withEmulatorNoZiip(cp int, ram int64) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }

  emulator = {
    cp  = %d
    ram = %d
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		cp,
		ram,
	)
}

func testAccInstanceResourceConfig_withZosCreds(username, password string) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }

  zos_creds = {
    username = "%s"
    password = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		username,
		password,
	)
}

func testAccInstanceResourceConfig_withIPL(deviceAddress, iodfAddress, loadSuffix string) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }

  ipl = {
    device_address = "%s"
    iodf_address   = "%s"
    load_suffix    = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		os.Getenv("INSTANCE_SSH_PASSWORD"),
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
		deviceAddress,
		iodfAddress,
		loadSuffix,
	)
}

func testAccInstanceResourceConfig_withSSHKey(keyContent string) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_key_file = %q

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		keyContent,
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
	)
}

func testAccInstanceResourceConfig_withSSHKeyAndPassphrase(keyContent, passphrase string) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user       = "%s"
  ssh_target_key_file   = %q
  ssh_target_passphrase = "%s"

  general = {
    label       = "%s"
    target_uuid = "%s"
    image_uuid  = "%s"
  }
}
`,
		os.Getenv("INSTANCE_SSH_USER"),
		keyContent,
		passphrase,
		os.Getenv("INSTANCE_LABEL"),
		os.Getenv("INSTANCE_TARGET_UUID"),
		os.Getenv("INSTANCE_IMAGE_UUID"),
	)
}

//nolint:all
func assertNotNullPath(resourceName string, path tfjsonpath.Path) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.NotNull())
}

func assertStringExactPath(resourceName string, path tfjsonpath.Path, expected string) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.StringExact(expected))
}

func assertInt64ExactPath(resourceName string, path tfjsonpath.Path, expected int64) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.Int64Exact(expected))
}

func testAccCheckInstanceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		provisionUUID := rs.Primary.ID
		if provisionUUID == "" {
			return fmt.Errorf("no ID is set")
		}

		// TODO: Add actual API call to delete the instance outside of Terraform
		// For now, this is a placeholder. In real implementation, you would:
		// 1. Get the ODE client from the provider
		// 2. Call client.Instance.Delete() directly
		// This simulates the resource being deleted outside of Terraform's control

		return fmt.Errorf("testAccCheckInstanceDisappears not fully implemented - needs ODE client access")
	}
}
