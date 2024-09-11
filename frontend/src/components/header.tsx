import React from "react";
import Image from "next/image";
import logo from "../../public/static/images/logo.svg";

const Header = () => {
  return (
    <header className="flex justify-between items-center p-4">
      <div className="m-4">
        <h3 className="scroll-m-20 text-2xl font-semibold tracking-tight">
          Externí služby
        </h3>

        <p>Monitorování stavu API našich externích partnerů</p>
      </div>

      <div className="pr-2">
        <Image src={logo} alt="Direct Technologies" width={100} height={58} />
      </div>
    </header>
  );
};

export default Header;
