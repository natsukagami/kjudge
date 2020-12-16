module.exports = {
    content: ["html/**/*.html", "html/*.html", "ts/**/*.ts", "ts/**/*.tsx"],
    safelist: [/^hljs.*/],
    defaultExtractor: (content) => content.match(/[\w-/:]+(?<!:)/g) || [],
};
