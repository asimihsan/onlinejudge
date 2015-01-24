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

    function createEditor(editorElement) {
        window.editor = CodeMirror.fromTextArea(editorElement, {});
        window.editor.setValue(localStorage[getPersistedTextKey()] || window.editor.getValue());
        window.editor.on("change", onChange);

    }
    function setupEditorSpaces() {
        window.editor.setOption("extraKeys", {
           Tab: function(cm) {
                var spaces = Array(cm.getOption("indentUnit") + 1).join(" ");
                cm.replaceSelection(spaces);
            }
        });
    }
    function setupJava(editorElement, text) {
        window.editor.setOption("mode", "text/x-java");
        window.editor.setOption("lineNumbers", true);
        window.editor.setOption("lineWrapping", true);
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        if (text) {
            window.editor.setValue(text);
        }
        setupEditorSpaces();
    }
    function setupPython(editorElement, text) {
        window.editor.setOption("mode", "python");
        window.editor.setOption("lineNumbers", true);
        window.editor.setOption("lineWrapping", true);
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        if (text) {
            window.editor.setValue(text);
        }
        setupEditorSpaces();
    }
    function setupRuby(editorElement, text) {
        window.editor.setOption("mode", "ruby");
        window.editor.setOption("lineNumbers", true);
        window.editor.setOption("lineWrapping", true);
        window.editor.setOption("tabSize", 2);
        window.editor.setOption("indentUnit", 2);
        if (text) {
            window.editor.setValue(text);
        }
        setupEditorSpaces();
    }

    function onLanguageSelected(editorElement, language, text) {
        switch(language) {
            case 'java':
                setupJava(editorElement, text);
                language = 'java';
                break;
            case 'python':
                setupPython(editorElement, text);
                language = 'python';
                break;
            case 'ruby':
                setupRuby(editorElement, text);
                language = 'ruby';
                break;
        }
        localStorage[getPersistedLanguageKey()] = language;
    }

    $(function() {
        var editorElement = document.getElementById("editor");
        createEditor(editorElement);

        $(".language-select").chosen({
            width: '30%'
        }).change(function(_, target) {
            var text = window.editor.getValue();
            language = target.selected;
            onLanguageSelected(editorElement, language, text);
        });

        var language = localStorage[getPersistedLanguageKey()] || 'python';
        $(".language-select").val(language);
        $(".language-select").trigger("chosen:updated");

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
        
            $.ajax({
                type: "POST",
                url: "http://www.runsomecode.com/run/" + language,
                data: window.editor.getValue(),
                contentType: "text/plain; charset=utf-8",
                success: function(response) {
                    $("#output").text(response);
                    l.stop();
                },
                failure: function(response) {
                    $("#output").text(response);
                    l.stop();
                }
            });
        });
    });
}());
