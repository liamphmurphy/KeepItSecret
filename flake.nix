{
  description = "KeepItSecret development environment and API build";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.gnumake
              pkgs.go_1_27
              pkgs.gopls
              pkgs.gotools
              pkgs.golangci-lint
            ];

            shellHook = ''
              export GOTOOLCHAIN=local
              export GOPATH="''${GOPATH:-$PWD/.cache/go}"
              export GOMODCACHE="$PWD/.cache/go/pkg/mod"
              export GOCACHE="$PWD/.cache/go/build"
              echo "KeepItSecret development shell"
              go version
            '';
          };
        });

      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          api = pkgs.buildGo127Module {
            pname = "keepitsecret-api";
            version = "0.1.0";
            src = ./api;
            vendorHash = null;
            subPackages = [ "cmd/api" ];
            ldflags = [ "-s" "-w" ];
            postInstall = ''
              mv "$out/bin/api" "$out/bin/keepitsecret-api"
            '';
          };

          default = self.packages.${system}.api;
        });
    };
}
