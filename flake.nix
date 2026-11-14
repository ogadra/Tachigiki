{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          f (
            import nixpkgs {
              inherit system;
              # Terraform は BSL ライセンスで unfree 扱い。allowUnfree = true だと他の unfree パッケージも黙って通るので、terraform だけを許可する
              config.allowUnfreePredicate = pkg: nixpkgs.lib.getName pkg == "terraform";
            }
          )
        );
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            ffmpeg
            wrangler
            terraform
          ];
        };
      });
    };
}
