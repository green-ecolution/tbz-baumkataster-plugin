{
  description = "Development enviroment for green ecolution tbz baumkataster plugin";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  } @ inputs: (flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
        pre-commit-check = inputs.pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            golangci-lint.enable = false;
            gofmt.enable = true;
            gotest.enable = false;
          };
        };
      in {
        # devShells."x86_64-linux".default = import ./shell.nix { inherit pkgs; };
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go_1_23
            air
            delve
            golangci-lint
            sqlite
            postgresql
            pkg-config

            yarn
            nodejs_22
            typescript

            gnumake
            git
            proj
            gitflow
            yq-go
          ];

          shellHook = ''
            go mod download
            ${pre-commit-check.shellHook}
          '';
        };
      }
    ));
}
