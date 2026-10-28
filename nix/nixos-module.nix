self:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.scurrytail;
  yamlFormat = pkgs.formats.yaml { };

  opts = import ./options.nix {
    inherit lib pkgs self;
    configTarget = "/etc/scurrytail/config.yaml";
  };

  configData = opts.buildConfigData cfg;
  configFile = yamlFormat.generate "scurrytail-config.yaml" configData;
in
{
  options.services.scurrytail = opts.options;

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    environment.etc."scurrytail/config.yaml" = lib.mkIf (configData != { }) {
      source = configFile;
    };

    systemd.user.services.scurrytail = {
      description = "ScurryTail - Nimble Linux system tray kernel error monitor";
      documentation = [ "https://github.com/comicdeed/scurrytail" ];
      partOf = [ "graphical-session.target" ];
      after = [ "graphical-session.target" ];

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.package}/bin/scurrytail";
        Restart = "on-failure";
        RestartSec = "5s";
        MemoryMax = "64M";
        LockPersonality = true;
        RestrictRealtime = true;
      };

      wantedBy = [ "graphical-session.target" ];
    };
  };
}
