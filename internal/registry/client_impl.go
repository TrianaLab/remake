package registry

import (
	"context"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TrianaLab/remake/config"
)

// DefaultClient interacts with an OCI registry using ORAS
type DefaultClient struct {
	cfg *config.Config
}

// NewDefaultClient returns a new registry Client configured by cfg
func NewDefaultClient(cfg *config.Config) Client {
	return &DefaultClient{cfg: cfg}
}

// Login authenticates against cfg.DefaultRegistry, then saves credentials on success
func (c *DefaultClient) Login(ctx context.Context, registry, user, pass string) error {
	reg, err := remote.NewRegistry(registry)
	if err != nil {
		return err
	}
	client := &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(registry, auth.Credential{Username: user, Password: pass}),
	}
	reg.Client = client
	if err := reg.Ping(ctx); err != nil {
		return err
	}
	// persistir credenciales solo al autenticar correctamente
	key := config.NormalizeKey(registry)
	viper.Set("registries."+key+".username", user)
	viper.Set("registries."+key+".password", pass)
	return viper.WriteConfig()
}

// Push uploads an artifact at path to the given reference
func (c *DefaultClient) Push(ctx context.Context, reference, path string) error {
	def := c.cfg.DefaultRegistry
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(def))
	if err != nil {
		return err
	}
	repoRef := ref.Context()
	repo, err := remote.NewRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return err
	}
	// aplicar credenciales si existen
	key := config.NormalizeKey(repoRef.RegistryStr())
	user := viper.GetString("registries." + key + ".username")
	pass := viper.GetString("registries." + key + ".password")
	if user != "" {
		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(repoRef.RegistryStr(), auth.Credential{Username: user, Password: pass}),
		}
	}
	fs, err := file.New(path)
	if err != nil {
		return err
	}
	defer fs.Close()
	if _, err := oras.Copy(ctx, fs, ref.Identifier(), repo, ref.Identifier(), oras.DefaultCopyOptions); err != nil {
		return err
	}
	return nil
}

// Pull downloads an artifact from the reference into dest directory
func (c *DefaultClient) Pull(ctx context.Context, reference, dest string) error {
	def := c.cfg.DefaultRegistry
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(def))
	if err != nil {
		return err
	}
	repoRef := ref.Context()
	repo, err := remote.NewRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return err
	}
	// aplicar credenciales si existen
	key := config.NormalizeKey(repoRef.RegistryStr())
	user := viper.GetString("registries." + key + ".username")
	pass := viper.GetString("registries." + key + ".password")
	if user != "" {
		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(repoRef.RegistryStr(), auth.Credential{Username: user, Password: pass}),
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	fs, err := file.New(dest)
	if err != nil {
		return err
	}
	defer fs.Close()
	if _, err := oras.Copy(ctx, repo, ref.Identifier(), fs, ref.Identifier(), oras.DefaultCopyOptions); err != nil {
		return err
	}
	return nil
}
