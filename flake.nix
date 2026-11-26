{
  # More info on nix shell
  # https://nixos-and-flakes.thiscute.world/development/intro#creating-a-development-environment

  description = "A Nix-flake-based Go + Vue 3 + Wails3 development environment";

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
            govulncheck

            # INFO: Wails
            webkitgtk_6_0
            gtk4
            mesa
            pkg-config
            x264
            gst_all_1.gstreamer # coreelements
            gst_all_1.gst-plugins-base # appsink
            gst_all_1.gst-plugins-good # v4l2src, pulsesrc
            gst_all_1.gst-plugins-bad # WebVTT
            pipewire

            typescript-language-server
            vscode-langservers-extracted
            vue-language-server
            prettierd
            # nodejs
            pnpm
          ];
          shellHook = ''
            export CGO_CFLAGS="-U_FORTIFY_SOURCE"
            export CGO_CXXFLAGS="-U_FORTIFY_SOURCE"
            export PATH=$PATH:~/go/bin

            export __EGL_VENDOR_LIBRARY_FILENAMES="${pkgs.mesa}/share/glvnd/egl_vendor.d/50_mesa.json"

            # Workaround for bug in webkit where app video doesn't rerender with v-if
            export WEBKIT_GST_ALLOW_PLAYBACK_OF_INVISIBLE_VIDEOS=1
          '';
        };
    };
}
