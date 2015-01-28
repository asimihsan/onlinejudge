/*
function grecaptchaOnLoad() {
    window.recaptcha = grecaptcha.render('recaptcha', {
        'sitekey': '6LcB8gATAAAAAN4SkOZ0o30BvUFq--YsNiPsIuWp',
    });
}
*/

(function() {
    function getPersistedTextKey() {
        return window.location.href + '::' + 'persistedText';
    }

    function getPersistedLanguageKey() {
        return window.location.href + '::' + 'persistedLanguage';
    }

    function onChange(editor) {
        localStorage[getPersistedTextKey()] = editor.getValue();
    }

    function setCommonEditorOptions() {
        window.editor.setOption("lineNumbers", true);
        window.editor.setOption("lineWrapping", true);
        window.editor.setOption("theme", "solarized");
        setupEditorSpaces();
    }

    function createEditor(editorElement) {
        window.editor = CodeMirror.fromTextArea(editorElement, {});
        window.editor.setValue(localStorage[getPersistedTextKey()] || window.editor.getValue());
        window.editor.on("change", onChange);
        setCommonEditorOptions();
    }
    function setupEditorSpaces() {
        window.editor.setOption("extraKeys", {
           Tab: function(cm) {
                var spaces = Array(cm.getOption("indentUnit") + 1).join(" ");
                cm.replaceSelection(spaces);
            }
        });
    }
    function setupC(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "text/x-csrc");
        if (text) {
            window.editor.setValue(text);
        }
    }
    function setupCPP(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "text/x-c++src");
        if (text) {
            window.editor.setValue(text);
        }
    }
    function setupJava(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "text/x-java");
        if (text) {
            window.editor.setValue(text);
        }
    }
    function setupPython(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "python");
        if (text) {
            window.editor.setValue(text);
        }
    }
    function setupRuby(editorElement, text) {
        window.editor.setOption("tabSize", 2);
        window.editor.setOption("indentUnit", 2);
        setCommonEditorOptions();
        window.editor.setOption("mode", "ruby");
        if (text) {
            window.editor.setValue(text);
        }
    }

    function onLanguageSelected(editorElement, language, text) {
        switch(language) {
            case 'c':
                setupC(editorElement, text);
                break;
            case 'cpp':
                setupCPP(editorElement, text);
                break;
            case 'java':
                setupJava(editorElement, text);
                break;
            case 'python':
                setupPython(editorElement, text);
                break;
            case 'ruby':
                setupRuby(editorElement, text);
                break;
        }
        localStorage[getPersistedLanguageKey()] = language;
    }

    //$(window).load(grecaptchaOnLoad);

    $(function() {
        var editorElement = document.getElementById("editor");
        createEditor(editorElement);

        $(".language-select").change(function(_, target) {
            var text = window.editor.getValue();
            language = $(".language-select").val();
            onLanguageSelected(editorElement, language, text);
        });

        // triggering chosen:updated doesn't trigger the callback function,
        // so we call it ourselves.
        var language = localStorage[getPersistedLanguageKey()] || 'python';
        $(".language-select").val(language);
        onLanguageSelected(editorElement, language, window.editor.getValue());

        $(".clear-output-button").click(function() {
            $("#output").text("");
        });
        $(".clear-code-button").click(function() {
            window.editor.setValue("");
            window.editor.clearHistory();
        });
        $(".submit-button").click(function() {
            var l = Ladda.create(document.querySelector('.submit-button'));
            l.start();
            l.isLoading();

            data = {
                'code': window.editor.getValue(),
                //'recaptcha': grecaptcha.getResponse(window.recaptcha),
            }
            $.ajax({
                type: "POST",
                url: "/run/" + language,
                data: JSON.stringify(data),
                contentType: "application/json; charset=utf-8",
                success: function(response) {
                    console.log(response);
                    $("#output").text(response.output);
                    l.stop();
                },
                failure: function(response) {
                    console.log(response);
                    $("#output").text(response.output);
                    l.stop();
                }
            });
        });
    });
}());
