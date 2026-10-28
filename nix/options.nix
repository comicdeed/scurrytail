{ lib, pkgs, self, configTarget ? "config.yaml" }:

{
  options = {
    enable = lib.mkEnableOption "ScurryTail - Nimble Linux system tray applet for tailing fast-moving kernel logs";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.system}.default or (pkgs.callPackage ./package.nix { });
      defaultText = lib.literalExpression "inputs.scurrytail.packages.\${pkgs.system}.default";
      description = "The scurrytail package to install and run.";
    };

    threshold = lib.mkOption {
      type = lib.types.nullOr lib.types.int;
      default = null;
      example = 5;
      description = "Number of errors within sliding window before alerting.";
    };

    window = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "1m";
      description = "Evaluation window duration (e.g. 1m, 30s).";
    };

    cooldown = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "5m";
      description = "Quiet period after notification before alerting again.";
    };

    backend = lib.mkOption {
      type = lib.types.nullOr (lib.types.enum [ "auto" "journalctl" "dmesg" ]);
      default = null;
      example = "journalctl";
      description = "Kernel log streaming backend (auto, journalctl, or dmesg).";
    };

    viewerCmd = lib.mkOption {
      type = lib.types.nullOr (lib.types.listOf lib.types.str);
      default = null;
      example = [ "ghostty" "-e" "journalctl" "-k" "-f" "-n" "50" ];
      description = "Command line array executed when clicking 'View Kernel Logs' in tray.";
    };

    settings = lib.mkOption {
      type = lib.types.attrsOf lib.types.anything;
      default = { };
      example = {
        backend = "dmesg";
        threshold = 3;
      };
      description = "Arbitrary settings serialized directly to ${configTarget}.";
    };
  };

  buildConfigData = cfg:
    lib.filterAttrs (_: v: v != null)
      {
        threshold = cfg.threshold;
        window = cfg.window;
        cooldown = cfg.cooldown;
        backend = cfg.backend;
        viewer_cmd = cfg.viewerCmd;
      } // cfg.settings;
}
