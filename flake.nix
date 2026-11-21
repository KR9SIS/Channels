{
  # More info on nix shell
  # https://nixos-and-flakes.thiscute.world/development/intro#creating-a-development-environment

  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixos-unstable/nixexprs.tar.xz";
  };

  outputs =
    { nixpkgs, ... }:
    let
      # system should match the system you are running on
      system = "x86_64-linux";
    in
    {
      devShells."${system}".default =
        let
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.mkShell {
          packages = with pkgs; [
            bashInteractive
            go
            gofumpt
            gopls
            delve

            # INFO: Wails
            webkitgtk_6_0
            gtk4
            pkg-config
            gcc

            gnumake
            mbake
            openssl

            typescript-language-server
            vscode-langservers-extracted
            vue-language-server
            prettierd
            nodejs
            pnpm
          ];
          shellHook = ''
            export CGO_CFLAGS="-U_FORTIFY_SOURCE"
            export CGO_CXXFLAGS="-U_FORTIFY_SOURCE"
            export PATH=$PATH:~/go/bin
          '';
        };
    };
}
