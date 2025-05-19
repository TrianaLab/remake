# 🚀 Remake

**Remake** is a powerful CLI tool that lets you wrap Makefiles as OCI artifacts, resolve remote includes, and run them effortlessly. Built on [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper), Remake makes it easy to package, share, and execute your Makefiles anywhere! 😎

---

## 🔍 Features

* 📦 **Run** local or remote Makefiles with caching support
* 🌐 **Pull** HTTP or OCI-hosted Makefiles and print their contents
* 📤 **Push** Makefiles as OCI artifacts to any registry
* 🔒 **Login** to OCI registries and store credentials securely
* 📋 **Version** command to print current Remake version
* ⚙️ Configurable defaults via `~/.remake/config.yaml`
* 🔄 Smart cache handling to speed up repeated fetches

---

## ⚙️ Installation

1. **Build from source**
```bash
   go install github.com/TrianaLab/remake
```

2. **Or download a pre-built binary** from the [Releases page](https://github.com/TrianaLab/remake/releases).

---

## 📖 Usage

All commands share a common syntax:

```
remake [command] [flags] [arguments]
```

### ▶️ run

Run a Makefile target (local or remote):

```
remake run -f Makefile build
remake run https://example.com/Makefile test
```

* `-f, --file`: path to Makefile or remote reference
* `--no-cache`: disable cache for this run

---

### 📥 pull

Fetch an artifact (HTTP or OCI) and print its content:

```
remake pull https://example.com/Makefile
remake pull oci://registry.example.com/myrepo:latest
```

* `--no-cache`: force re-download, bypassing cache

---

### 📤 push

Package a Makefile as an OCI artifact and push to registry:

```
remake push oci://registry.example.com/myrepo:1.0.0 -f Makefile
```

* `-f, --file`: Makefile to push (defaults to `Makefile`)
* Uses Viper-configured credentials for authentication

---

### 🔐 login

Authenticate to an OCI registry and save credentials:

```
remake login registry.example.com
Username: <your-username>
Password: <your-password>
```

* `-u, --username` and `-p, --password` flags supported

---

### 📋 version

Print the current version:

```
remake version
# → Version: 1.2.3
```

---

## 🛠️ Configuration

By default, Remake stores its config in:

```
~/.remake/config.yaml
```

Default settings:

```
cacheDir: ~/.remake/cache
defaultMakefile: makefile
insecure: false
registries: {}
```

You can edit this file to change:

* `cacheDir`: where downloaded files are cached
* `defaultMakefile`: default file name when running or pushing
* `insecure`: allow HTTP (`true`)
* `registries`: map of registry endpoints to credentials

---

## 🤝 Contributing

We welcome contributions! Feel free to open issues or submit pull requests:

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/foo`)
3. Commit your changes (`git commit -am 'Add feature'`)
4. Push to branch (`git push origin feature/foo`)
5. Open a Pull Request 🚀

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).
