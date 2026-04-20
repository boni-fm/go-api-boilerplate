// Package handlers defines the HTTP handler layer for the application.
//
// # Dependency Philosophy
//
// Handlers depend directly on concrete service structs (e.g. *services.UserService)
// rather than on service-level interfaces. This keeps the dependency graph
// immediately readable for new engineers — there is no need to hunt for which
// concrete type implements a given interface.
//
// Testability is preserved at the repository layer: inject a mock
// repository.UserRepository into services.NewUserService when constructing
// test fixtures. See handlers/testing_helpers.go and handlers/user_test.go.
package handlers
