# home-manager module: per-user systemd service + timer.
self:
{ config, lib, pkgs, ... }:
let
  cfg = config.services.cloudflare-dynamic-dns;
  r = import ./render.nix { inherit lib pkgs cfg; };
in
{
  imports = [ (import ./options.nix self) ];

  config = lib.mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.domains != [ ];
        message = "services.cloudflare-dynamic-dns.domains must not be empty.";
      }
      {
        assertion = (cfg.iface != null) != (cfg.ipcmd != null);
        message = "services.cloudflare-dynamic-dns: exactly one of `iface` or `ipcmd` must be set.";
      }
      {
        assertion = !cfg.multihost || cfg.hostId != null;
        message = "services.cloudflare-dynamic-dns.hostId is required when multihost is enabled.";
      }
    ];

    systemd.user.services.cloudflare-dynamic-dns = {
      Unit.Description = "Dynamic DNS for Cloudflare";
      Service = {
        Type = "oneshot";
        ExecStart = r.execStart;
        LoadCredential = "token:${cfg.tokenFile}";
      } // lib.optionalAttrs cfg.stateFile {
        StateDirectory = "cloudflare-dynamic-dns";
      };
    };

    systemd.user.timers.cloudflare-dynamic-dns = {
      Unit.Description = "Run cloudflare-dynamic-dns every ${cfg.interval}";
      Timer = {
        OnBootSec = "1min";
        OnUnitActiveSec = cfg.interval;
        Persistent = true;
      };
      Install.WantedBy = [ "timers.target" ];
    };
  };
}
