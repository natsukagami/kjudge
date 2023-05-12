const purgeCSS = require("@fullhuman/postcss-purgecss");

const purgecssConfig = require("./purgecss.config.js");

module.exports = {
    plugins: [
        "tailwindcss",
        purgeCSS(purgecssConfig)
    ],
};
