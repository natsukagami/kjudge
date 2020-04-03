module.exports = {
    content: ["html/**/*.html", "html/*.html", "ts/**/*.ts", "ts/**/*.tsx"],
    whitelist: [/^hljs.*/],
    defaultExtractor: (content) => content.match(/[\w-/:]+(?<!:)/g) || [],
};
