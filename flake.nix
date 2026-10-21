{
  description = "Lightweight Linux kernel error monitor daemon";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "armv7l-linux" "riscv64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor = system: import nixpkgs { inherit system; };
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = self.packages.${system}.dmesg-watch;
          dmesg-watch = pkgs.callPackage ./nix/package.nix { };
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.dmesg-watch}/bin/dmesg-watch";
          meta.description = "Run dmesg-watch directly via nix run";
        };
      });

      devShells = forAllSystems (system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              systemd
              util-linux
              xdg-utils
            ];
          };
        }
      );

      formatter = forAllSystems (system: (pkgsFor system).nixpkgs-fmt);

      overlays = {
        default = self.overlays.dmesg-watch;
        dmesg-watch = final: prev: {
          dmesg-watch = final.callPackage ./nix/package.nix { };
        };
      };

      nixosModules = {
        default = self.nixosModules.dmesg-watch;
        dmesg-watch = import ./nix/nixos-module.nix self;
      };

      homeManagerModules = {
        default = self.homeManagerModules.dmesg-watch;
        dmesg-watch = import ./nix/hm-module.nix self;
      };

      homeModules = self.homeManagerModules;
    };
}
