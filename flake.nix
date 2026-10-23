{
  description = "Recruitment system for hiring new employee";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
      };
      ems = pkgs.buildGoModule {
        pname = "backend";
        version = "1";

        src = ./backend;
	vendorHash = "sha256-3Z058vwtCWrMpWwVVGCACzDJU0WbcY0ObLmCu+8C7nY=";
	proxyVendor = true;
      };
      helper = pkgs.writeShellApplication {
        name = "ems-compile";
        runtimeInputs = [
          pkgs.go
        ];
        text = ''
          go build -C backend -o main
          mv backend/main ./main
        '';
      };
    in
    {
      packages = {
        ${system} = {
          default = pkgs.stdenv.mkDerivation {
            pname = "full system";
            version = "1";

            src = ./frontend;

            skipBuildPhase = true;
            installPhase = ''
              mkdir -p $out/share $out/bin
              cp -r $src $out/share/frontend
              cp -r ${ems}/bin/ems $out/bin/backend
            '';
          };
        };
      };
      devShells = {
        ${system} = {
          default = pkgs.mkShell {
            packages = [ helper pkgs.go ];
            shellHook = ''
              echo "welcome run helper command";
              echo "=> command name: ems-compile";
            '';
          };
        };
      };
    };
}
