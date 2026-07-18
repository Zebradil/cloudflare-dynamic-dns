# Shared option declarations for the NixOS and home-manager modules.
# `self` is this flake, used to default `package` to the flake's own build.
self:
{ lib, pkgs, ... }:
let
  inherit (lib) mkOption mkEnableOption types;
in
{
  options.services.cloudflare-dynamic-dns = {
    enable = mkEnableOption "the Cloudflare dynamic DNS client";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "cloudflare-dynamic-dns package from this flake";
      description = "The cloudflare-dynamic-dns package to run.";
    };

    domains = mkOption {
      type = types.listOf types.str;
      default = [ ];
      example = [ "example.com" "*.example.com" ];
      description = "Domain names to assign the detected address to.";
    };

    tokenFile = mkOption {
      # types.str (not types.path) so a runtime secret path is used as-is; a
      # path literal would be copied into the world-readable Nix store.
      type = types.str;
      example = "/run/secrets/cloudflare-dynamic-dns-token";
      description = ''
        Path to a file containing the Cloudflare API token. The token is passed
        to the service via systemd credentials and is never written to the
        world-readable Nix store. A trailing newline is stripped.
      '';
    };

    hostId = mkOption {
      type = types.nullOr types.str;
      default = null;
      description = "Unique host identifier. Required in multihost mode.";
    };

    iface = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "eth0";
      description = "Network interface to look up for an address. Mutually exclusive with `ipcmd`.";
    };

    ipcmd = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "curl -fsSL https://api6.ipify.org";
      description = "Shell command to run to get the address. Mutually exclusive with `iface`.";
    };

    logLevel = mkOption {
      type = types.enum [ "trace" "debug" "info" "warning" "error" "fatal" "panic" ];
      default = "info";
      description = "Logging level.";
    };

    multihost = mkOption {
      type = types.bool;
      default = false;
      description = "Enable multihost mode (several hosts sharing one domain).";
    };

    prioritySubnets = mkOption {
      type = types.listOf types.str;
      default = [ ];
      example = [ "2001:db8::/32" ];
      description = "Subnets to prefer when several addresses are found.";
    };

    proxy = mkOption {
      type = types.enum [ "auto" "enabled" "disabled" ];
      default = "auto";
      description = "Cloudflare proxy setting for created or updated records.";
    };

    stack = mkOption {
      type = types.enum [ "ipv4" "ipv6" ];
      default = "ipv6";
      description = "IP stack version.";
    };

    ttl = mkOption {
      type = types.int;
      default = 1;
      description = "DNS record TTL in seconds, or 1 for 'automatic'.";
    };

    stateFile = mkOption {
      type = types.bool;
      default = true;
      description = ''
        Keep a state file to skip Cloudflare API calls when the address is
        unchanged. Stored under the service's systemd StateDirectory.
      '';
    };

    interval = mkOption {
      type = types.str;
      default = "5m";
      example = "10m";
      description = "How often to run, as a systemd time span (timer OnUnitActiveSec).";
    };
  };
}
