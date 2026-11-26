{
  pkgs,
  self,
}:
let
  baseVersion = "4.5.16";
  commit = self.shortRev or self.dirtyShortRev or "unknown";
  version = "${baseVersion}-${commit}";
in
pkgs.buildGoModule {
  inherit version;
  pname = "cloudflare-dynamic-dns";
  src = ./..;
  vendorHash = "sha256-V0o6Zv8lKwDKzFvKdV37pxlHrZZcuzKRsv0kHtHpp2I=";

  env.CGO_ENABLED = 0;
  ldflags = [
    "-s"
    "-w"
    "-X=main.version=${baseVersion}"
    "-X=main.commit=${commit}"
    "-X=main.date=1970-01-01"
  ];

  meta = {
    changelog = "https://github.com/Zebradil/cloudflare-dynamic-dns/blob/${baseVersion}/CHANGELOG.md";
    description = "Dynamic DNS client for Cloudflare";
    homepage = "https://github.com/Zebradil/cloudflare-dynamic-dns";
    license = pkgs.lib.licenses.mit;
  };
}
