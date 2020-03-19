module.exports = {
    content: ["html/**/*.html", "html/*.html"],
    whitelist: [/^hljs.*/],
    defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || [],
};
