{ lib
, buildGo127Module
, makeWrapper
, systemd
, util-linux
, xdg-utils
}:

buildGo127Module {
  pname = "dmesg-watch";
  version = "0.3.0";

  src = lib.cleanSource ./..;

  vendorHash = "sha256-Xk0jyJsjtVCKsY4kJ5piLdcJ4a8+ZkyM7jTrnjGZnYA=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.3.0"
  ];

  nativeBuildInputs = [ makeWrapper ];

  postInstall = ''
    wrapProgram $out/bin/dmesg-watch \
      --prefix PATH : ${lib.makeBinPath [ systemd util-linux xdg-utils ]}
  '';

  meta = with lib; {
    description = "Lightweight Linux kernel error monitor with desktop notifications and tray applet";
    homepage = "https://github.com/codemedic/dmesg-watch";
    license = licenses.mit;
    mainProgram = "dmesg-watch";
    platforms = platforms.linux;
  };
}
