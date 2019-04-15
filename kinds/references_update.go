package kinds

import (
	"context"

	"github.com/creativesoftwarefdn/weaviate/database/schema"
	"github.com/creativesoftwarefdn/weaviate/database/schema/kind"
	"github.com/creativesoftwarefdn/weaviate/models"
	"github.com/creativesoftwarefdn/weaviate/validation"
	"github.com/go-openapi/strfmt"
)

// UpdateActionReferences Class Instance to the connected DB. If the class contains a network
// ref, it has a side-effect on the schema: The schema will be updated to
// include this particular network ref class.
func (m *Manager) UpdateActionReferences(ctx context.Context, id strfmt.UUID,
	propertyName string, refs models.MultipleRef) error {
	schemaLock, err := m.db.SchemaLock()
	if err != nil {
		return newErrInternal("could not aquire lock: %v", err)
	}
	defer unlock(schemaLock)
	classSchema := schemaLock.GetSchema()
	schemaManager := schemaLock.SchemaManager()
	dbConnector := schemaLock.Connector()

	return m.updateActionReferenceToConnectorAndSchema(ctx, id, propertyName, refs,
		dbConnector, classSchema, schemaManager)
}

func (m *Manager) updateActionReferenceToConnectorAndSchema(ctx context.Context, id strfmt.UUID,
	propertyName string, refs models.MultipleRef, repo updateRepo, classSchema schema.Schema,
	schemaManager schemaManager) error {

	// get action to see if it exists
	action, err := m.getActionFromRepo(ctx, id, repo)
	if err != nil {
		return err
	}

	err = m.validateReferences(ctx, refs, repo)
	if err != nil {
		return err
	}

	err = m.validateCanModifyReference(kind.ACTION_KIND, action.Class, propertyName, classSchema)
	if err != nil {
		return err
	}

	updatedSchema, err := m.replaceClassPropReferences(action.Schema, propertyName, refs)
	if err != nil {
		return err
	}
	action.Schema = updatedSchema
	action.LastUpdateTimeUnix = unixNow()

	// the new refs could be network refs
	err = m.addNetworkDataTypesForAction(ctx, schemaManager, action)
	if err != nil {
		return newErrInternal("could not update schema for network refs: %v", err)
	}

	repo.UpdateAction(ctx, action, action.ID)
	if err != nil {
		return newErrInternal("could not store action: %v", err)
	}

	return nil
}

// UpdateThingReferences Class Instance to the connected DB. If the class contains a network
// ref, it has a side-effect on the schema: The schema will be updated to
// include this particular network ref class.
func (m *Manager) UpdateThingReferences(ctx context.Context, id strfmt.UUID,
	propertyName string, refs models.MultipleRef) error {
	schemaLock, err := m.db.SchemaLock()
	if err != nil {
		return newErrInternal("could not aquire lock: %v", err)
	}
	defer unlock(schemaLock)
	classSchema := schemaLock.GetSchema()
	schemaManager := schemaLock.SchemaManager()
	dbConnector := schemaLock.Connector()

	return m.updateThingReferenceToConnectorAndSchema(ctx, id, propertyName, refs,
		dbConnector, classSchema, schemaManager)
}

func (m *Manager) updateThingReferenceToConnectorAndSchema(ctx context.Context, id strfmt.UUID,
	propertyName string, refs models.MultipleRef, repo updateRepo, classSchema schema.Schema,
	schemaManager schemaManager) error {

	// get thing to see if it exists
	thing, err := m.getThingFromRepo(ctx, id, repo)
	if err != nil {
		return err
	}

	err = m.validateReferences(ctx, refs, repo)
	if err != nil {
		return err
	}

	err = m.validateCanModifyReference(kind.THING_KIND, thing.Class, propertyName, classSchema)
	if err != nil {
		return err
	}

	updatedSchema, err := m.replaceClassPropReferences(thing.Schema, propertyName, refs)
	if err != nil {
		return err
	}
	thing.Schema = updatedSchema
	thing.LastUpdateTimeUnix = unixNow()

	// the new refs could be network refs
	err = m.addNetworkDataTypesForThing(ctx, schemaManager, thing)
	if err != nil {
		return newErrInternal("could not update schema for network refs: %v", err)
	}

	repo.UpdateThing(ctx, thing, thing.ID)
	if err != nil {
		return newErrInternal("could not store thing: %v", err)
	}

	return nil
}

func (m *Manager) validateReferences(ctx context.Context, references models.MultipleRef, repo getRepo) error {
	err := validation.ValidateMultipleRef(ctx, m.config, &references, repo, m.network, "reference not found")
	if err != nil {
		return newErrInvalidUserInput("invalid references: %v", err)
	}

	return nil
}

func (m *Manager) replaceClassPropReferences(props interface{}, propertyName string,
	refs models.MultipleRef) (interface{}, error) {

	if props == nil {
		props = map[string]interface{}{}
	}

	propsMap := props.(map[string]interface{})
	propsMap[propertyName] = refs
	return propsMap, nil
}