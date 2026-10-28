self:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.scurrytail;
  yamlFormat = pkgs.formats.yaml { };

  opts = import ./options.nix {
    inherit lib pkgs self;
    configTarget = "~/.config/scurrytail/config.yaml";
  };

  configData = opts.buildConfigData cfg;
  configFile = yamlFormat.generate "scurrytail-config.yaml" configData;
in
{
  options.services.scurrytail = opts.options;

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];

    xdg.configFile."scurrytail/config.yaml" = lib.mkIf (configData != { }) {
      source = configFile;
    };

    systemd.user.services.scurrytail = {
      Unit = {
        Description = "ScurryTail - Nimble Linux system tray kernel error monitor";
        Documentation = [ "https://github.com/comicdeed/scurrytail" ];
        PartOf = [ "graphical-session.target" ];
        After = [ "graphical-session.target" ];
      };

      Service = {
        Type = "simple";
        ExecStart = "${cfg.package}/bin/scurrytail";
        Restart = "on-failure";
        RestartSec = "5s";
        MemoryMax = "64M";
        LockPersonality = true;
        RestrictRealtime = true;
      };

      Install = {
        WantedBy = [ "graphical-session.target" ];
      };
    };
  };
}
