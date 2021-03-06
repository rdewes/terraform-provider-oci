// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/oracle/oci-go-sdk/common"
	oci_file_storage "github.com/oracle/oci-go-sdk/filestorage"
)

var (
	snapshotDataSourceRepresentation = map[string]interface{}{
		"file_system_id": Representation{repType: Required, create: `${oci_file_storage_file_system.test_file_system.id}`},
		"id":             Representation{repType: Optional, create: `${oci_file_storage_snapshot.test_snapshot.id}`},
		"state":          Representation{repType: Optional, create: `ACTIVE`},
		"filter":         RepresentationGroup{Required, snapshotDataSourceFilterRepresentation}}
	snapshotDataSourceFilterRepresentation = map[string]interface{}{
		"name":   Representation{repType: Required, create: `id`},
		"values": Representation{repType: Required, create: []string{`${oci_file_storage_snapshot.test_snapshot.id}`}},
	}

	snapshotRepresentation = map[string]interface{}{
		"file_system_id": Representation{repType: Required, create: `${oci_file_storage_file_system.test_file_system.id}`},
		"name":           Representation{repType: Required, create: `snapshot-1`},
	}

	SnapshotResourceDependencies = AvailabilityDomainConfig + FileSystemRequiredOnlyResource
)

func TestFileStorageSnapshotResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getEnvSettingWithBlankDefault("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)

	resourceName := "oci_file_storage_snapshot.test_snapshot"
	datasourceName := "data.oci_file_storage_snapshots.test_snapshots"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		CheckDestroy: testAccCheckFileStorageSnapshotDestroy,
		Steps: []resource.TestStep{
			// verify create
			{
				Config: config + compartmentIdVariableStr + SnapshotResourceDependencies +
					generateResourceFromRepresentationMap("oci_file_storage_snapshot", "test_snapshot", Required, Create, snapshotRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "file_system_id"),
					resource.TestCheckResourceAttr(resourceName, "name", "snapshot-1"),
				),
			},

			// verify datasource
			{
				Config: config +
					generateDataSourceFromRepresentationMap("oci_file_storage_snapshots", "test_snapshots", Optional, Update, snapshotDataSourceRepresentation) +
					compartmentIdVariableStr + SnapshotResourceDependencies +
					generateResourceFromRepresentationMap("oci_file_storage_snapshot", "test_snapshot", Optional, Update, snapshotRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "file_system_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "state", "ACTIVE"),

					resource.TestCheckResourceAttr(datasourceName, "snapshots.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "snapshots.0.file_system_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "snapshots.0.id"),
					resource.TestCheckResourceAttr(datasourceName, "snapshots.0.name", "snapshot-1"),
					resource.TestCheckResourceAttrSet(datasourceName, "snapshots.0.state"),
					resource.TestCheckResourceAttrSet(datasourceName, "snapshots.0.time_created"),
				),
			},
			// verify resource import
			{
				Config:                  config,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ResourceName:            resourceName,
			},
		},
	})
}

func testAccCheckFileStorageSnapshotDestroy(s *terraform.State) error {
	noResourceFound := true
	client := testAccProvider.Meta().(*OracleClients).fileStorageClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "oci_file_storage_snapshot" {
			noResourceFound = false
			request := oci_file_storage.GetSnapshotRequest{}

			tmp := rs.Primary.ID
			request.SnapshotId = &tmp

			response, err := client.GetSnapshot(context.Background(), request)

			if err == nil {
				deletedLifecycleStates := map[string]bool{
					string(oci_file_storage.SnapshotLifecycleStateDeleted): true,
				}
				if _, ok := deletedLifecycleStates[string(response.LifecycleState)]; !ok {
					//resource lifecycle state is not in expected deleted lifecycle states.
					return fmt.Errorf("resource lifecycle state: %s is not in expected deleted lifecycle states", response.LifecycleState)
				}
				//resource lifecycle state is in expected deleted lifecycle states. continue with next one.
				continue
			}

			//Verify that exception is for '404 not found'.
			if failure, isServiceError := common.IsServiceError(err); !isServiceError || failure.GetHTTPStatusCode() != 404 {
				return err
			}
		}
	}
	if noResourceFound {
		return fmt.Errorf("at least one resource was expected from the state file, but could not be found")
	}

	return nil
}
