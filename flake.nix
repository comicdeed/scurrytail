{
  description = "A nimble Linux system tray applet for tailing fast-moving kernel logs and catching errors in real time";

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
          default = self.packages.${system}.scurrytail;
          scurrytail = pkgs.callPackage ./nix/package.nix { };
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.scurrytail}/bin/scurrytail";
          meta.description = "Run scurrytail directly via nix run";
        };
      });

      devShells = forAllSystems (system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_27
              systemd
              util-linux
              xdg-utils
            ];
          };
        }
      );

      formatter = forAllSystems (system: (pkgsFor system).nixpkgs-fmt);

      overlays = {
        default = self.overlays.scurrytail;
        scurrytail = final: prev: {
          scurrytail = final.callPackage ./nix/package.nix { };
        };
      };

      nixosModules = {
        default = self.nixosModules.scurrytail;
        scurrytail = import ./nix/nixos-module.nix self;
      };

      homeManagerModules = {
        default = self.homeManagerModules.scurrytail;
        scurrytail = import ./nix/hm-module.nix self;
      };

      homeModules = self.homeManagerModules;
    };
}
