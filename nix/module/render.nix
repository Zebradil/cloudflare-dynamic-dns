# Renders module options into a store YAML config and a start wrapper.
# The token is intentionally absent from the config: it is read at runtime from
# the systemd credential and exported as CFDDNS_TOKEN.
{ lib, pkgs, cfg }:
let
  yamlFormat = pkgs.formats.yaml { };

  # CLI/YAML keys are kebab-case; drop unset (null) values.
  settings = lib.filterAttrs (_: v: v != null) {
    inherit (cfg) domains stack proxy ttl multihost iface ipcmd;
    "log-level" = cfg.logLevel;
    "priority-subnets" = cfg.prioritySubnets;
    "host-id" = cfg.hostId;
  };

  configFile = yamlFormat.generate "cloudflare-dynamic-dns.yaml" settings;

  # An empty --state-file enables the state file with an auto-derived path under
  # $STATE_DIRECTORY (matching the shipped systemd unit).
  args = [ "--config=${configFile}" ] ++ lib.optional cfg.stateFile "--state-file=";

  wrapper = pkgs.writeShellScript "cloudflare-dynamic-dns-start" ''
    set -eu
    export CFDDNS_TOKEN="$(<"$CREDENTIALS_DIRECTORY/token")"
    exec ${cfg.package}/bin/cloudflare-dynamic-dns ${lib.escapeShellArgs args}
  '';
in
{
  inherit configFile;
  execStart = "${wrapper}";
}
