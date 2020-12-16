const defaultTheme = require("tailwindcss/defaultTheme");
const colors = require("tailwindcss/colors");

module.exports = {
    purge: false,
    theme: {
        extend: {},
        fontFamily: {
            ...defaultTheme.fontFamily,
            mono: ['"IBM Plex Mono"', ...defaultTheme.fontFamily.mono],
        },
        colors: {
            ...defaultTheme.colors,
            teal: colors.teal,
            orange: colors.orange,
        },
    },
    variants: {},
    plugins: [],
};
