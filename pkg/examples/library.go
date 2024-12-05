import "github.com/SUSE/connect/collectors"
import "github.com/SUSE/connect/connection"
import "github.com/SUSE/connect/registration"
import "github.com/SUSE/connect/validation"

type ClientCredentials struct {
  user string
  pw string
  online_credentials string
  offline_key string
}

creds = ClientCredentials.New()


# generic informations
conn = connection.New().ToSCC()

info, err := connection.SubscriptionInfo(conn, "someregcode")
packages, err := connection.PackageSearch(conn, "some package")


# online registration + activation

# SCC
conn = connection.WithCredentials(creds).ToSCC()
# Proxy
conn = connection.WithCredentials(creds).ToRegistrationProxy(url)
# With custom proxy settings
conn = connection.WithCredentials(creds).WithCustomProxySettings(handler).ToSCC()


status, err := registration.Status(conn)
systemInformation, err = collectors.DefaultCollectors().Run()

switch status {
  case registration.Registered:
    println("I'm registered do nothing but ping and show activations")
    err = registration.Keepalive(systemInformation)

    activations, err = registration.Activations(conn)
    for activation := range(activations) {
      printf("activated product: %" + activation.name)
    }

  case registration.Unregistered:
    println("I'm unregistered registering now..")
    err = registration.Register(regcode)

    productTree, err = registration.Activate(conn, "SLES", "15.5", "x86_64", systemInformation)

    for product := range(productTree.Flat(productTree.WithRecommended | productTree.NoMigrationExtras) {
      product, err = registration.Activate(conn, product.Identifier, product.Version, product.Architecture, nil)
    }
  }
}

# online deactivation
conn = connection.WithCredentials(creds).ToSCC()
err = registration.Deactivate(conn, "sle-container-module", "15.5", "x86_64")


# online deregistration
conn = connection.WithCredentials(creds).ToSCC()
err = registration.Deregister(conn)


# offline validation
validator = validator.PayloadFromReader(reader)
status, err = validator.Validate()

switch status {
  case validator.Invalid:
    printf("Invalid payload supplied")
  case validator.Valid:
    info, err = validator.SubscriptionInfo()
    printf("You use subscription %s", info.Name)
}
