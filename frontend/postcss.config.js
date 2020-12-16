const purgecss = require("@fullhuman/postcss-purgecss");
const purgecssConfig = require("./purgecss.config.js");

module.exports = {
    plugins: [
        // ...
        require("tailwindcss")(require("./tailwind.config.js")),
        require("autoprefixer"),
        purgecss(purgecssConfig),
        // ...
    ],
};
