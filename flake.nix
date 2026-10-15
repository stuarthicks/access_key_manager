{
  description = "access_key_manager - CLI for rotating AWS Access Keys";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAll = f: nixpkgs.lib.genAttrs systems (s: f nixpkgs.legacyPackages.${s});
    in {
      packages = forAll (pkgs: rec {
        access_key_manager = pkgs.buildGoModule {
          pname = "access_key_manager";
          version = self.shortRev or self.dirtyShortRev or "dev";
          src = ./.;
          vendorHash = "sha256-Jp8JAR+wnZ9reDKWRsMcXqbiIENwpbtqjDtyo/4+Nbs=";
          ldflags = [ "-s" "-w" ];
          meta.mainProgram = "access_key_manager";
        };
        default = access_key_manager;
      });

      devShells = forAll (pkgs: {
        default = pkgs.mkShell { packages = [ pkgs.go pkgs.just pkgs.goreleaser ]; };
      });
    };
}
