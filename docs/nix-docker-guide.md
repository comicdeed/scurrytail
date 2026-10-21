# Nix Development via Docker Guide

When developing on a host where the native `nix` toolchain is not installed, all Nix operations can be performed directly through Docker using the official `nixos/nix` container image.

---

## 🚀 Makefile Targets (Dynamic Detection)

The project `Makefile` automatically detects whether `nix` is installed on your host. If `nix` is missing, it transparently routes commands through Docker:

| Make Target | Description |
| :--- | :--- |
| **`make nix-build`** | Compiles the package derivation (`nix build path:.`). |
| **`make nix-check`** | Validates flake syntax, derivations, dev shells, and NixOS/HM modules (`nix flake check path:. --no-build`). |
| **`make nix-fmt`** | Formats all `.nix` files using `nixpkgs-fmt`. |

---

## 🐳 Docker Invocation Formula

The underlying command formula mounts the repository workspace into `nixos/nix:latest` and enables Flakes:

```bash
docker run --rm \
  -v "$PWD:$PWD" \
  -w "$PWD" \
  nixos/nix:latest \
  nix --extra-experimental-features "nix-command flakes" <COMMAND>
```

### Examples

**1. Build the package derivation:**
```bash
docker run --rm -v "$PWD:$PWD" -w "$PWD" nixos/nix:latest \
  nix --extra-experimental-features "nix-command flakes" build path:.
```

**2. Evaluate Flake schema and module definitions:**
```bash
docker run --rm -v "$PWD:$PWD" -w "$PWD" nixos/nix:latest \
  nix --extra-experimental-features "nix-command flakes" flake check path:. --no-build
```

**3. Format Nix files:**
```bash
docker run --rm -v "$PWD:$PWD" -w "$PWD" nixos/nix:latest \
  nix --extra-experimental-features "nix-command flakes" run nixpkgs#nixpkgs-fmt -- flake.nix nix/*.nix
```

---

## 🔄 Updating `vendorHash`

When Go dependencies in `go.mod` change:

1. In [`nix/package.nix`](../nix/package.nix), set:
   ```nix
   vendorHash = lib.fakeHash;
   ```
2. Run `make nix-build`.
3. Nix will fail with a hash mismatch error displaying the computed hash:
   ```text
   specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
   got:       sha256-l1gTwnm/lksn/UnLp7iRELFmjpIche3IOjIzYCw+6T0=
   ```
4. Update `vendorHash` in [`nix/package.nix`](../nix/package.nix) with the `got` value.
5. Re-run `make nix-build` to verify the build completes successfully.
