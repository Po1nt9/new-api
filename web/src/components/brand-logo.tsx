/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import React from "react";

import { DEFAULT_LOGO, DEFAULT_LOGO_DARK } from "@/lib/constants";
import { cn } from "@/lib/utils";

export interface BrandLogoProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src: string;
}

/**
 * Brand logo image with theme-aware variants.
 *
 * When `src` is the default logo, renders both the light variant (black
 * spark) and the dark variant (white spark) and swaps them with `dark:`
 * classes, mirroring the waffo logo pattern. Any other `src` (e.g. a logo
 * URL configured in system settings) renders as a single image.
 */
export function BrandLogo({
  src,
  alt = "logo",
  className,
  ...props
}: BrandLogoProps) {
  if (src && src !== DEFAULT_LOGO) {
    return <img src={src} alt={alt} className={className} {...props} />;
  }

  return (
    <>
      <img
        src={DEFAULT_LOGO}
        alt={alt}
        className={cn(className, "dark:hidden")}
        {...props}
      />
      <img
        src={DEFAULT_LOGO_DARK}
        alt={alt}
        className={cn(className, "hidden dark:block")}
        {...props}
      />
    </>
  );
}
