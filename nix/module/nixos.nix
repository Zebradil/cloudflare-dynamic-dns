# NixOS module: system-wide systemd service + timer.
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

    systemd.services.cloudflare-dynamic-dns = {
      description = "Dynamic DNS for Cloudflare";
      after = [ "network-online.target" "nss-lookup.target" ];
      wants = [ "network-online.target" ];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = r.execStart;
        LoadCredential = [ "token:${cfg.tokenFile}" ];

        # Hardening.
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        # AF_NETLINK is needed to enumerate interface addresses.
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" "AF_NETLINK" ];
        RestrictNamespaces = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
      } // lib.optionalAttrs cfg.stateFile {
        StateDirectory = "cloudflare-dynamic-dns";
      };
    };

    systemd.timers.cloudflare-dynamic-dns = {
      description = "Run cloudflare-dynamic-dns every ${cfg.interval}";
      wantedBy = [ "timers.target" ];
      timerConfig = {
        OnBootSec = "1min";
        OnUnitActiveSec = cfg.interval;
        Persistent = true;
      };
    };
  };
}
