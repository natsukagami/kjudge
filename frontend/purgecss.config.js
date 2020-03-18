module.exports = {
    content: ["html/**/*.html", "html/*.html"],
    whitelist: [],
    defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || [],
};
