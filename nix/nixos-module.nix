self:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.dmesg-watch;
  yamlFormat = pkgs.formats.yaml { };

  configData = lib.filterAttrs (_: v: v != null)
    {
      threshold = cfg.threshold;
      window = cfg.window;
      cooldown = cfg.cooldown;
      backend = cfg.backend;
      viewer_cmd = cfg.viewerCmd;
    } // cfg.settings;

  configFile = yamlFormat.generate "dmesg-watch-config.yaml" configData;
in
{
  options.services.dmesg-watch = {
    enable = lib.mkEnableOption "Lightweight Linux kernel error monitor daemon";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.system}.default or (pkgs.callPackage ./package.nix { });
      defaultText = lib.literalExpression "inputs.dmesg-watch.packages.\${pkgs.system}.default";
      description = "The dmesg-watch package to install and run.";
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
      description = "Arbitrary settings serialized directly to /etc/dmesg-watch/config.yaml.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    environment.etc."dmesg-watch/config.yaml" = lib.mkIf (configData != { }) {
      source = configFile;
    };

    systemd.user.services.dmesg-watch = {
      description = "Lightweight Linux kernel error monitor daemon";
      documentation = [ "https://github.com/codemedic/dmesg-watch" ];
      partOf = [ "graphical-session.target" ];
      after = [ "graphical-session.target" ];

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.package}/bin/dmesg-watch";
        Restart = "on-failure";
        RestartSec = "5s";
      };

      wantedBy = [ "graphical-session.target" ];
    };
  };
}
