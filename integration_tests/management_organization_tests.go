package integration_tests

import (
	"context"
	"testing"

	_ "github.com/kernelplex/ubase/internal/evercoregen"
	"github.com/kernelplex/ubase/lib/ubmanage"
	"github.com/kernelplex/ubase/lib/ubstatus"
)

var anOrganization = ubmanage.OrganizationState{
	Name:       "Test Organization",
	SystemName: "test-organization",
	Status:     "active",
}

var updatedOrganization = ubmanage.OrganizationState{
	Name:       "Updated Organization",
	SystemName: "updated-organization",
	Status:     "inactive",
}

func (s *ManagmentServiceTestSuite) AddOrganization(t *testing.T) {
	response, err := s.managementService.OrganizationAdd(context.Background(), ubmanage.OrganizationCreateCommand{
		Name:       anOrganization.Name,
		SystemName: anOrganization.SystemName,
		Status:     anOrganization.Status,
	}, "test-runner")

	if err != nil {
		t.Fatalf("AddOrganization failed to add organization: %v", err)
	}

	s.createdOrganizationId = response.Data.Id
}

func (s *ManagmentServiceTestSuite) GetOrganizationBySystemName(t *testing.T) {

	res, err := s.managementService.OrganizationGetBySystemName(context.Background(), anOrganization.SystemName)
	if err != nil {
		t.Fatalf("GetOrganizationBySystemName failed to get organization by system name: %v", err)
	}

	if res.Status != ubstatus.Success {
		t.Fatalf("GetOrganizationBySystemName status is not success: %v", res.Status)
	}
	if res.Data.State.Name != anOrganization.Name {
		t.Fatalf("GetOrganizationBySystemName organization name does not match: %v", res.Data.State.Name)
	}

	if res.Data.State.SystemName != anOrganization.SystemName {
		t.Fatalf("GetOrganizationBySystemName system name does not match: %v", res.Data.State.SystemName)
	}
}

func (s *ManagmentServiceTestSuite) UpdateOrganization(t *testing.T) {
	res, err := s.managementService.OrganizationUpdate(context.Background(), ubmanage.OrganizationUpdateCommand{
		Id:         s.createdOrganizationId,
		Name:       &updatedOrganization.Name,
		SystemName: &updatedOrganization.SystemName,
		Status:     &updatedOrganization.Status,
	}, "test-runner")

	if err != nil {
		t.Fatalf("UpdateOrganization failed to update organization: %v", err)
	}
	if res.Status != ubstatus.Success {
		t.Fatalf("UpdateOrganization status is not success: %v", res.Status)
	}

	response, err := s.managementService.OrganizationGetBySystemName(context.Background(), updatedOrganization.SystemName)
	if err != nil {
		t.Fatalf("UpdateOrganization failed to get organization by system name: %v", err)
	}

	if response.Status != ubstatus.Success {
		t.Fatalf("UpdateOrganization status is not success: %v", response.Status)
	}

	retrievedOrganization := response.Data.State
	if retrievedOrganization.Name != updatedOrganization.Name {
		t.Fatalf("UpdateOrganization organization name does not match: %v", retrievedOrganization.Name)
	}

	if retrievedOrganization.SystemName != updatedOrganization.SystemName {
		t.Fatalf("UpdateOrganization system name does not match: %v", retrievedOrganization.SystemName)
	}
	if retrievedOrganization.Status != updatedOrganization.Status {
		t.Fatalf("UpdateOrganization status does not match: %v", retrievedOrganization.Status)
	}
	if retrievedOrganization.Status != updatedOrganization.Status {
		t.Fatalf("UpdateOrganization status does not match: %v", retrievedOrganization.Status)
	}
}
