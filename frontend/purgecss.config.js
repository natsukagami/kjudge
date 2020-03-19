module.exports = {
    content: ["html/**/*.html", "html/*.html", "ts/*.ts"],
    whitelist: [/^hljs.*/],
    defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || [],
};
