// @ts-check
import colors from "tailwindcss/colors";
import defaultTheme from "tailwindcss/defaultTheme";

/** @type {import("tailwindcss").Config} */
const tailwindConfig = {
    content: ["./html/**/*.html", "./ts/**/*.{ts,tsx}"],
    theme: {
        extend: {
            fontFamily: {
                mono: ['"IBM Plex Mono"', ...defaultTheme.fontFamily.mono],
            },
            colors: {
                teal: colors.teal,
                orange: colors.orange,
            },
        },
    },
    plugins: [],
};

export default tailwindConfig;
