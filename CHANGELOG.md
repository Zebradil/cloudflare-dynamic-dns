## [4.0.4](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/4.0.3...4.0.4) (2024-03-02)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.89.0 ([400cfed](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/400cfed76b86d25323779d2565e50e56fb17303e))

## [4.0.3](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/4.0.2...4.0.3) (2024-02-14)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.88.0 ([f96afbf](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/f96afbf2a942ea5618207787433b3a3bd62bd075))

## [4.0.2](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/4.0.1...4.0.2) (2024-01-31)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.87.0 ([66a8fd8](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/66a8fd83e2affa94d40be5f1b98614628e72f057))

## [4.0.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/4.0.0...4.0.1) (2024-01-17)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.86.0 ([0fce146](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/0fce14601be25394a43b48bf3cfeeef15f2108a5))

# [4.0.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/3.1.0...4.0.0) (2024-01-08)


### Bug Fixes

* consistently use kebab-case for config keys ([429d6a9](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/429d6a972a95d9f27e527a1c920e14dfb63e0e6c))


### BREAKING CHANGES

* the `prioritySubnets` config key has been renamed to
`priority-subnets` to be consistent with the other config keys.

# [3.1.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/3.0.0...3.1.0) (2024-01-07)


### Features

* rename with-state-file flag to state-file, fix systemd unit ([27803f9](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/27803f90cfb184580b28b6805c0995139f51146c))

# [3.0.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.8.1...3.0.0) (2024-01-07)


### Bug Fixes

* add check for STATE_DIRECTORY when running in systemd mode ([25ef6de](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/25ef6de6c4e59a1e1793382e10945cdca1239626))
* copy only binary in the container image ([d82e524](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/d82e524fe91adb7161bec60b702a3ea7857e4a62))
* do not shadow variable via assignment ([19c7d5a](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/19c7d5a577d518f47e66375eed5b89ed56ca9c45))
* switching between single- and multi-host modes ([70e39af](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/70e39af2290466ec3747a84a85bb3a46bb0bdf21))
* validate token and iface values ([3ab4e44](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/3ab4e4488da1a479f71fc9d1baf3052201bf75b5))


### Features

* add --run-every to run periodically ([ac2af3c](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/ac2af3c7f558b267ef06bc0d1648f6996fa87f60))
* include "managed by" in the DNS record comment ([d829f28](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/d829f282770d9cd1b8fc1ff9cb39b79230f359f3))
* replace --systemd flag with --with-state-file ([adbc8df](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/adbc8df70ee216d89fcfa035a7d8b20a8cb979a6))


### BREAKING CHANGES

* --systemd flag is replaced with --with-state-file.

## [2.8.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.8.0...2.8.1) (2024-01-07)


### Bug Fixes

* set path to config file via envvar ([4623ac7](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/4623ac74a303f84c2e133d84ec25d26355f77e81))

# [2.8.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.7.0...2.8.0) (2024-01-07)


### Features

* accept environment variables for configuration ([249d9d8](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/249d9d88aa105c3e6fc15b801c212be39c883985))

# [2.7.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.6.1...2.7.0) (2024-01-07)


### Features

* add experimental multihost mode ([c5bcb0a](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/c5bcb0a45ab9797140863c379080afbfd898ea9c))

## [2.6.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.6.0...2.6.1) (2024-01-06)


### Bug Fixes

* improve EUI-64 detection by checking for the FFFE injection ([fd4da5b](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/fd4da5b0e98005c54d62fc847d543babc7ff68a8))
* use 7th bit from the _left_ to identify EUI-64 addresses ([a144a4c](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/a144a4c2dfc7cc25d86bb9d11ae0a8d8b69d2cec))

# [2.6.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.5.1...2.6.0) (2024-01-05)


### Features

* add --version flag and log ([8fe1e51](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/8fe1e515ab8d8d0f6ea6ff8f023ff519eced6713))

## [2.5.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.5.0...2.5.1) (2024-01-03)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.85.0 ([24f1bb7](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/24f1bb7cec073b5279f383ccad491eda64645197))

# [2.5.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.4.0...2.5.0) (2024-01-01)


### Bug Fixes

* log a warning if the selected IPv6 address is not optimal ([0a13ac3](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/0a13ac3f59f099cf9788cb3c544e6d583500a0f9))


### Features

* prefer EUI-64 over randomly generated identifiers ([3d5d772](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/3d5d772d3a57c2984a87bf9b4b9b25b40a31df30))

# [2.4.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.9...2.4.0) (2024-01-01)


### Features

* prefer GUA over ULA when selecting IPv6 address ([fa8e21a](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/fa8e21aa1d41038a3f8f72da3145a8dcc33f2ce9))

## [2.3.9](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.8...2.3.9) (2024-01-01)


### Bug Fixes

* add ca-certificates to scratch image ([9579a11](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/9579a116874e922921f6137d1be916802e73d4fa)), closes [#90](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/90)

## [2.3.8](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.7...2.3.8) (2023-12-20)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.84.0 ([3cc6ae6](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/3cc6ae653ce8ddc42355a002826b10f6a8f4b508))

## [2.3.7](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.6...2.3.7) (2023-12-18)


### Bug Fixes

* **deps:** update module github.com/spf13/viper to v1.18.2 ([fb3084b](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/fb3084b8a43a2ad1728f2b927c167217115a2309))

## [2.3.6](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.5...2.3.6) (2023-12-08)


### Bug Fixes

* **deps:** update module github.com/spf13/viper to v1.18.1 ([4f85b3b](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/4f85b3bcef1400b633a6ba9509746cab41622178))

## [2.3.5](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.4...2.3.5) (2023-12-06)


### Bug Fixes

* **deps:** update module github.com/spf13/viper to v1.18.0 ([f18c747](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/f18c747e8c38e878c50ee3e001b63dfce6baaa69))

## [2.3.4](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.3...2.3.4) (2023-12-06)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.83.0 ([1700071](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/1700071de0c8ff37510f30d3677d5d87236656c6))

## [2.3.3](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.2...2.3.3) (2023-11-22)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.82.0 ([9a3ad02](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/9a3ad023b12b02b609398dc130df4c1c2487c059))

## [2.3.2](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.1...2.3.2) (2023-11-08)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.81.0 ([fe49f28](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/fe49f28c52cca21fe9afdbdfbeea8fe8a83d2983))

## [2.3.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.3.0...2.3.1) (2023-11-06)


### Bug Fixes

* **ci:** adjust goreleaser config for less docker arches ([bbcfd76](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/bbcfd76585fbf3f67a463bb03b0f6fe056bfe8df))

# [2.3.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.8...2.3.0) (2023-11-06)


### Bug Fixes

* **ci:** log in to ghcr.io ([3cb4299](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/3cb42996145e473e82cf14352f5a214cd995a7a0))
* **release:** install systemd units and create config dir in AUR package ([64d90df](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/64d90df804e52f743ce5e0b5d1c0a4a1ce3e5078))
* **release:** reduce docker build targets ([272f7df](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/272f7df3cfe218d3d7f0806c56a885e48949dae9))
* **release:** render goreleaser config ([b88e551](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/b88e55169f685061f515b732a684253f9df38029))


### Features

* **release:** add binary AUR package ([fc765be](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/fc765bece8f07c214e9ebbdbc9cc9a89ea3e80c0))
* **release:** add nfpms packages configuration ([a05c01b](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/a05c01bc7e4a74f2bed076bc40f4a37d9540f64e))

## [2.2.8](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.7...2.2.8) (2023-11-04)


### Bug Fixes

* **deps:** update module github.com/spf13/cobra to v1.8.0 ([48517b6](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/48517b6033b109b4095d70fb119214dc2ac1d63c))

## [2.2.7](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.6...2.2.7) (2023-10-25)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.80.0 ([d8b2975](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/d8b29753f280cf9c108c9d259bdd201234380df8))

## [2.2.6](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.5...2.2.6) (2023-10-11)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.79.0 ([6ec4248](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/6ec42488dd4152eaeafc8b95c8e80aa3f8bb7a7f))

## [2.2.5](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.4...2.2.5) (2023-10-06)


### Bug Fixes

* **deps:** update module github.com/spf13/viper to v1.17.0 ([ca0d63c](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/ca0d63c3d8df248f761bf20c454d089b83cac6d0))

## [2.2.4](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.3...2.2.4) (2023-09-27)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.78.0 ([bca9b6b](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/bca9b6bfd99803f579efef4b0c35e93d38efc39c))

## [2.2.3](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.2...2.2.3) (2023-09-13)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.77.0 ([08b3e46](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/08b3e4606915f22db458eef8192b6e4a389d6569))

## [2.2.2](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.1...2.2.2) (2023-08-30)


### Bug Fixes

* **ci:** configure more platforms for Docker ([e27660f](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/e27660f1ad97bca0879e825a2e7aa9062b85b591))

## [2.2.1](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.2.0...2.2.1) (2023-08-30)


### Bug Fixes

* **ci:** trigger release ([5f74e40](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/5f74e40747c7bcdde73da1f9ad365aa219894397))

# [2.2.0](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.1.3...2.2.0) (2023-08-30)


### Features

* add Docker support ([#64](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/64)) ([f67188a](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/f67188ae08fccbc6b02eda31bc975b3b0ed76efa))

## [2.1.3](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.1.2...2.1.3) (2023-08-30)


### Bug Fixes

* **deps:** update module github.com/cloudflare/cloudflare-go to v0.76.0 ([a260354](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/a260354b4af4998d94fe3a44094bffb4fe71cc9c))

## [2.1.2](https://github.com/Zebradil/cloudflare-dynamic-dns/compare/2.1.1...2.1.2) (2023-08-23)


### Bug Fixes

* check returned error ([613a78f](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/613a78fde6d0e4f0995de5d9274f916dcddc9c51))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.65.0 ([1823074](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/182307434f89e6a81d7240fd8df5d7707ddf972f))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.66.0 ([409738a](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/409738a5846202c6a53cd7ef5bfe50846a6d66f5))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.68.0 ([89d3496](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/89d3496999abe68f1c5791060746b49fa20e8a74))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.69.0 ([#55](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/55)) ([07bd6aa](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/07bd6aae333b22ad2ee5a45117361891674d5bee))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.70.0 ([#56](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/56)) ([039fe31](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/039fe31514d54dc1571ac461fb6a9982141e2424))
* **deps:** update module github.com/cloudflare/cloudflare-go to v0.75.0 ([#57](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/57)) ([0acef0e](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/0acef0ea1b1c7a82d923751759169800b7fd6f96))
* **deps:** update module github.com/sirupsen/logrus to v1.9.2 ([28110bb](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/28110bb659b2b768c4b99373f459b5bc44d88e4a))
* **deps:** update module github.com/sirupsen/logrus to v1.9.3 ([#54](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/54)) ([bd1d917](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/bd1d91706f7e0f731ab8e272f32b8ccabae77f4d))
* **deps:** update module github.com/spf13/viper to v1.16.0 ([#53](https://github.com/Zebradil/cloudflare-dynamic-dns/issues/53)) ([bd77422](https://github.com/Zebradil/cloudflare-dynamic-dns/commit/bd77422b3d3a672af6094c0b52cfdb69ce8d7e80))
