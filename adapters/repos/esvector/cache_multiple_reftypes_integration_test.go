// +build integrationTest

package esvector

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v5"
	"github.com/go-openapi/strfmt"
	"github.com/semi-technologies/weaviate/adapters/handlers/graphql/local/get"
	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/schema"
	"github.com/semi-technologies/weaviate/entities/schema/kind"
	"github.com/semi-technologies/weaviate/usecases/traverser"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultipleCrossRefTypes(t *testing.T) {
	// This test suite has various scenarios for a class with multiple ref types
	// in different caching scenarios. For all tests we want to make sure that
	// all types can be retrieved and that only those are retrieved that the user
	// asked for (select properties)

	t.Run("one level deep", func(t *testing.T) {

		refSchema := schema.Schema{
			Things: &models.Schema{
				Classes: []*models.Class{
					&models.Class{
						Class: "MultiRefParkingGarage",
						Properties: []*models.Property{
							&models.Property{
								Name:     "name",
								DataType: []string{string(schema.DataTypeString)},
							},
						},
					},
					&models.Class{
						Class: "MultiRefParkingLot",
						Properties: []*models.Property{
							&models.Property{
								Name:     "name",
								DataType: []string{string(schema.DataTypeString)},
							},
						},
					},
					&models.Class{
						Class: "MultiRefCar",
						Properties: []*models.Property{
							&models.Property{
								Name:     "name",
								DataType: []string{string(schema.DataTypeString)},
							},
							&models.Property{
								Name:     "parkedAt",
								DataType: []string{"MultiRefParkingGarage", "MultiRefParkingLot"},
							},
						},
					},
				},
			},
		}

		client, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{"http://localhost:9201"},
		})
		require.Nil(t, err)

		logger, _ := test.NewNullLogger()
		schemaGetter := &fakeSchemaGetter{schema: refSchema}
		repo := NewRepo(client, logger, schemaGetter, 2)
		waitForEsToBeReady(t, repo)
		requestCounter := &testCounter{}
		repo.requestCounter = requestCounter
		migrator := NewMigrator(repo)

		t.Run("adding all classes to the schema", func(t *testing.T) {
			for _, class := range refSchema.Things.Classes {
				t.Run(fmt.Sprintf("add %s", class.Class), func(t *testing.T) {
					err := migrator.AddClass(context.Background(), kind.Thing, class)
					require.Nil(t, err)
				})
			}
		})

		t.Run("importing with various combinations of props", func(t *testing.T) {
			objects := []models.Thing{
				models.Thing{
					Class: "MultiRefParkingGarage",
					Schema: map[string]interface{}{
						"name": "Luxury Parking Garage",
					},
					ID:               "a7e10b55-1ac4-464f-80df-82508eea1951",
					CreationTimeUnix: 1566469890,
				},
				models.Thing{
					Class: "MultiRefParkingGarage",
					Schema: map[string]interface{}{
						"name": "Crappy Parking Garage",
					},
					ID:               "ba2232cf-bb0e-413d-b986-6aa996d34d2e",
					CreationTimeUnix: 1566469892,
				},
				models.Thing{
					Class: "MultiRefParkingLot",
					Schema: map[string]interface{}{
						"name": "Fancy Parking Lot",
					},
					ID:               "1023967b-9512-475b-8ef9-673a110b695d",
					CreationTimeUnix: 1566469894,
				},
				models.Thing{
					Class: "MultiRefParkingLot",
					Schema: map[string]interface{}{
						"name": "The worst parking lot youve ever seen",
					},
					ID:               "901859d8-69bf-444c-bf43-498963d798d2",
					CreationTimeUnix: 1566469897,
				},
				models.Thing{
					Class: "MultiRefCar",
					Schema: map[string]interface{}{
						"name": "Car which is parked no where",
					},
					ID:               "329c306b-c912-4ec7-9b1d-55e5e0ca8dea",
					CreationTimeUnix: 1566469899,
				},
				models.Thing{
					Class: "MultiRefCar",
					Schema: map[string]interface{}{
						"name": "Car which is parked in a garage",
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
						},
					},
					ID:               "fe3ca25d-8734-4ede-9a81-bc1ed8c3ea43",
					CreationTimeUnix: 1566469902,
				},
				models.Thing{
					Class: "MultiRefCar",
					Schema: map[string]interface{}{
						"name": "Car which is parked in a lot",
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					},
					ID:               "21ab5130-627a-4268-baef-1a516bd6cad4",
					CreationTimeUnix: 1566469906,
				},
				models.Thing{
					Class: "MultiRefCar",
					Schema: map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					},
					ID:               "533673a7-2a5c-4e1c-b35d-a3809deabace",
					CreationTimeUnix: 1566469909,
				},
			}

			for _, thing := range objects {
				t.Run(fmt.Sprintf("add %s", thing.ID), func(t *testing.T) {
					err := repo.PutThing(context.Background(), &thing, []float32{1, 2, 3, 4, 5, 6, 7})
					require.Nil(t, err)
				})
			}

			refreshAll(t, client)

			t.Run("before cache indexing", func(t *testing.T) {
				t.Run("car with no refs", func(t *testing.T) {
					var id strfmt.UUID = "329c306b-c912-4ec7-9b1d-55e5e0ca8dea"
					expectedSchema := map[string]interface{}{
						"name": "Car which is parked no where",
						"uuid": id.String(),
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})
				})

				t.Run("car with single ref to garage", func(t *testing.T) {
					var id strfmt.UUID = "fe3ca25d-8734-4ede-9a81-bc1ed8c3ea43"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
						},
					}

					expectedSchemaNoRefs := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						// ref is not present at all
					}

					expectedSchemaWithRefs := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingGarage",
								Fields: map[string]interface{}{
									"name": "Luxury Parking Garage",
									"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
								},
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 2, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 2, requestCounter.count)
						assert.Equal(t, expectedSchemaNoRefs, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 3, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})
				})

				t.Run("car with single ref to lot", func(t *testing.T) {
					var id strfmt.UUID = "21ab5130-627a-4268-baef-1a516bd6cad4"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					}

					expectedSchemaNoRefs := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						// ref is not present at all
					}

					expectedSchemaWithRefs := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingLot",
								Fields: map[string]interface{}{
									"name": "Fancy Parking Lot",
									"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
								},
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 2, requestCounter.count)
						assert.Equal(t, expectedSchemaNoRefs, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 2, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 3, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})
				})

				t.Run("car with refs to both", func(t *testing.T) {
					var id strfmt.UUID = "533673a7-2a5c-4e1c-b35d-a3809deabace"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					}

					expectedSchemaWithLotRef := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingLot",
								Fields: map[string]interface{}{
									"name": "Fancy Parking Lot",
									"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
								},
							},
						},
					}
					expectedSchemaWithGarageRef := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingGarage",
								Fields: map[string]interface{}{
									"name": "Luxury Parking Garage",
									"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
								},
							},
						},
					}
					expectedSchemaWithAllRefs := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingLot",
								Fields: map[string]interface{}{
									"name": "Fancy Parking Lot",
									"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
								},
							},
							get.LocalRef{
								Class: "MultiRefParkingGarage",
								Fields: map[string]interface{}{
									"name": "Luxury Parking Garage",
									"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
								},
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 3, requestCounter.count)
						assert.Equal(t, expectedSchemaWithGarageRef, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 3, requestCounter.count)
						assert.Equal(t, expectedSchemaWithLotRef, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 5, requestCounter.count)
						assert.Equal(t, expectedSchemaWithAllRefs, res.Schema)
					})
				})

			})

			repo.InitCacheIndexing(50, 200*time.Millisecond, 200*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)

			t.Run("verify that cache is hot", func(t *testing.T) {
				// by checking if the cache of the last imported thing is hot

				res, err := repo.ThingByID(context.Background(), "fe3ca25d-8734-4ede-9a81-bc1ed8c3ea43", nil)
				require.Nil(t, err)
				require.Equal(t, true, res.CacheHot)
			})

			t.Run("after cache indexing", func(t *testing.T) {
				t.Run("car with no refs", func(t *testing.T) {
					var id strfmt.UUID = "329c306b-c912-4ec7-9b1d-55e5e0ca8dea"
					expectedSchema := map[string]interface{}{
						"name": "Car which is parked no where",
						"uuid": id.String(),
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchema, res.Schema)
					})
				})

				t.Run("car with single ref to garage", func(t *testing.T) {
					var id strfmt.UUID = "fe3ca25d-8734-4ede-9a81-bc1ed8c3ea43"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
						},
					}

					expectedSchemaNoRefs := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						// ref is not present at all
					}

					expectedSchemaWithRefs := map[string]interface{}{
						"name": "Car which is parked in a garage",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingGarage",
								Fields: map[string]interface{}{
									"name": "Luxury Parking Garage",
									"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
								},
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaNoRefs, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})
				})

				t.Run("car with single ref to lot", func(t *testing.T) {
					var id strfmt.UUID = "21ab5130-627a-4268-baef-1a516bd6cad4"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					}

					expectedSchemaNoRefs := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						// ref is not present at all
					}

					expectedSchemaWithRefs := map[string]interface{}{
						"name": "Car which is parked in a lot",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingLot",
								Fields: map[string]interface{}{
									"name": "Fancy Parking Lot",
									"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
								},
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaNoRefs, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithRefs, res.Schema)
					})
				})

				t.Run("car with refs to both", func(t *testing.T) {
					var id strfmt.UUID = "533673a7-2a5c-4e1c-b35d-a3809deabace"
					expectedSchemaUnresolved := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						// ref is present, but unresolved, therefore the lowercase letter
						"parkedAt": models.MultipleRef{
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/a7e10b55-1ac4-464f-80df-82508eea1951",
							},
							&models.SingleRef{
								Beacon: "weaviate://localhost/things/1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
					}

					expectedSchemaWithLotRef := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingLot",
								Fields: map[string]interface{}{
									"name": "Fancy Parking Lot",
									"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
								},
							},
						},
					}
					expectedSchemaWithGarageRef := map[string]interface{}{
						"name": "Car which is parked in two places at the same time (magic!)",
						"uuid": id.String(),
						"ParkedAt": []interface{}{
							get.LocalRef{
								Class: "MultiRefParkingGarage",
								Fields: map[string]interface{}{
									"name": "Luxury Parking Garage",
									"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
								},
							},
						},
					}

					bothRefs := []interface{}{
						get.LocalRef{
							Class: "MultiRefParkingLot",
							Fields: map[string]interface{}{
								"name": "Fancy Parking Lot",
								"uuid": "1023967b-9512-475b-8ef9-673a110b695d",
							},
						},
						get.LocalRef{
							Class: "MultiRefParkingGarage",
							Fields: map[string]interface{}{
								"name": "Luxury Parking Garage",
								"uuid": "a7e10b55-1ac4-464f-80df-82508eea1951",
							},
						},
					}

					t.Run("asking for no refs", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, nil)
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaUnresolved, res.Schema)
					})

					t.Run("asking for refs of type garage", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtGarage())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithGarageRef, res.Schema)
					})

					t.Run("asking for refs of type lot", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtLot())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						assert.Equal(t, expectedSchemaWithLotRef, res.Schema)
					})

					t.Run("asking for refs of both types", func(t *testing.T) {
						requestCounter.reset()
						res, err := repo.ThingByID(context.Background(), id, parkedAtEither())
						require.Nil(t, err)

						assert.Equal(t, 1, requestCounter.count)
						parkedAt, ok := res.Schema.(map[string]interface{})["ParkedAt"]
						require.True(t, ok)

						parkedAtSlice, ok := parkedAt.([]interface{})
						require.True(t, ok)

						assert.ElementsMatch(t, bothRefs, parkedAtSlice)
					})
				})

			})

		})
	})
}

func parkedAtGarage() traverser.SelectProperties {
	return traverser.SelectProperties{
		traverser.SelectProperty{
			Name:        "ParkedAt",
			IsPrimitive: false,
			Refs: []traverser.SelectClass{
				traverser.SelectClass{
					ClassName: "MultiRefParkingGarage",
					RefProperties: traverser.SelectProperties{
						traverser.SelectProperty{
							Name:        "name",
							IsPrimitive: true,
						},
					},
				},
			},
		},
	}
}

func parkedAtLot() traverser.SelectProperties {
	return traverser.SelectProperties{
		traverser.SelectProperty{
			Name:        "ParkedAt",
			IsPrimitive: false,
			Refs: []traverser.SelectClass{
				traverser.SelectClass{
					ClassName: "MultiRefParkingLot",
					RefProperties: traverser.SelectProperties{
						traverser.SelectProperty{
							Name:        "name",
							IsPrimitive: true,
						},
					},
				},
			},
		},
	}
}

func parkedAtEither() traverser.SelectProperties {
	return traverser.SelectProperties{
		traverser.SelectProperty{
			Name:        "ParkedAt",
			IsPrimitive: false,
			Refs: []traverser.SelectClass{
				traverser.SelectClass{
					ClassName: "MultiRefParkingLot",
					RefProperties: traverser.SelectProperties{
						traverser.SelectProperty{
							Name:        "name",
							IsPrimitive: true,
						},
					},
				},
				traverser.SelectClass{
					ClassName: "MultiRefParkingGarage",
					RefProperties: traverser.SelectProperties{
						traverser.SelectProperty{
							Name:        "name",
							IsPrimitive: true,
						},
					},
				},
			},
		},
	}
}