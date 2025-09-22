package permifysync

import (
	"context"
	"errors"
	"strings"

	permissions "bitbucket.org/up2metricPC/u2m-permissions"
	organizationPermissions "bitbucket.org/up2metricPC/u2m-permissions/entities/organization"
	kafkadispatcher "github.com/Nerzal/gocloak/v13/pkg/kafka-dispatcher"
)

type PermifyHandler struct {
	permifyClient *permissions.AuthzClient
	dispatcher    *kafkadispatcher.Dispatcher
	realm         string
}

func NewPermifyHandler(permifyClient *permissions.AuthzClient, dispatcher *kafkadispatcher.Dispatcher, realm string) kafkadispatcher.Handler {
	return &PermifyHandler{
		permifyClient: permifyClient,
		dispatcher:    dispatcher,
		realm:         realm,
	}
}

func (h *PermifyHandler) Handle(ctx context.Context, event kafkadispatcher.Event) error {
	if h.realm != event.RealmName {
		return nil
	}

	switch event.ResourceType {
	case kafkadispatcher.ResourceOrganizationMembership:
		if event.OperationType.Equal(kafkadispatcher.OperationTypeCreate) {
			return h.handleAddUser(ctx, event)
		} else if event.OperationType.Equal(kafkadispatcher.OperationTypeDelete) {
			return h.handleRemoveUser(ctx, event)
		}

	case kafkadispatcher.ResourceOrganization:
		if event.OperationType.Equal(kafkadispatcher.OperationTypeCreate) {
			return h.handleCreateOrganization(ctx, event)
		} else if event.OperationType.Equal(kafkadispatcher.OperationTypeDelete) {
			return h.handleDeleteOrganization(ctx, event)
		}
	}

	return nil

}

func (h *PermifyHandler) handleAddUser(_ context.Context, event kafkadispatcher.Event) error {
	// TODO:This might not always work actually
	tokens := strings.Split(string(event.ResourcePath), "/")
	if len(tokens) != 4 || tokens[0] != "organizations" || tokens[2] != "members" {
		return errors.New("invalid resource path")
	}

	orgID := tokens[1]
	userID := tokens[3]

	if err := organizationPermissions.AddUser(h.permifyClient, orgID, userID); err != nil {
		return err
	}

	return nil
}

func (h *PermifyHandler) handleRemoveUser(_ context.Context, event kafkadispatcher.Event) error {
	// TODO: This might not always work actually
	tokens := strings.Split(string(event.ResourcePath), "/")
	if len(tokens) != 4 || tokens[0] != "organizations" || tokens[2] != "members" {
		return errors.New("invalid resource path")
	}

	orgID := tokens[1]
	userID := tokens[3]

	if err := organizationPermissions.RemoveUser(h.permifyClient, orgID, userID); err != nil {
		return err
	}

	return nil
}

func (h *PermifyHandler) handleCreateOrganization(_ context.Context, event kafkadispatcher.Event) error {
	// TODO
	return nil
}

func (h *PermifyHandler) handleDeleteOrganization(_ context.Context, event kafkadispatcher.Event) error {
	orgID, err := event.ResourcePath.ExtractOrganizationID()
	if err != nil {
		return err
	}

	if err := organizationPermissions.DeleteOrganization(h.permifyClient, orgID); err != nil {
		return err
	}

	return nil
}
