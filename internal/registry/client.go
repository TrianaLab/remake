package registry

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TrianaLab/remake/config"
)

type Client interface {
	Login(ctx context.Context, registry, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) ([]byte, error)
}

type DefaultClient struct {
	cfg *config.Config
}

func New(cfg *config.Config) Client {
	return &DefaultClient{cfg: cfg}
}

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
	key := config.NormalizeKey(registry)
	viper.Set("registries."+key+".username", user)
	viper.Set("registries."+key+".password", pass)
	return viper.WriteConfig()
}

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

func (c *DefaultClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	def := c.cfg.DefaultRegistry
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(def))
	if err != nil {
		return nil, err
	}

	repoRef := ref.Context()
	repo, err := remote.NewRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return nil, err
	}

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

	store := memory.New()

	if _, err := oras.Copy(ctx, repo, ref.Identifier(), store, ref.Identifier(), oras.DefaultCopyOptions); err != nil {
		return nil, err
	}
	/*
		manifestDesc, err := store.Resolve(ctx, ref.Identifier())
		if err != nil {
			return nil, err
		}

		manifestBytes, err := content.FetchAll(ctx, store.Fetch(ctx, ref.Identifier()), manifestDesc)
		if err != nil {
			return nil, err
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return nil, err
		}
		if len(manifest.Layers) == 0 {
			return nil, fmt.Errorf("no layers found in artifact %s", reference)
		}

		layerDesc := manifest.Layers[0]
		data, err := content.FetchAll(ctx, store, layerDesc)
		if err != nil {
			return nil, err
		}

		return data, nil*/
	return nil, nil
}
