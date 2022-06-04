{ pkgs, lib, config, ... }:
let
  cfg = config.services.kjudge;
in
with lib;
{
  options.services.kjudge = {
    enable = mkEnableOption "Enable kjudge, a minimal programming contest judge.";
    port = mkOption {
      type = types.int;
      default = 8088;
      description = "The port to listen on";
    };
    adminKey = mkOption {
      type = types.nullOr types.str;
      default = null;
      description = "The admin key, for logging onto the Admin Panel. If null, an admin key will be generated every start up.";
    };

  };
}

