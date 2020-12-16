const defaultTheme = require("tailwindcss/defaultTheme");

module.exports = {
    purge: false,
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
