const defaultTheme = require("tailwindcss/defaultTheme");

module.exports = {
    theme: {
        extend: {},
        fontFamily: {
            ...defaultTheme.fontFamily,
            mono: ['"IBM Plex Mono"', ...defaultTheme.fontFamily.mono],
        },
    },
    variants: {},
    plugins: [],
};
