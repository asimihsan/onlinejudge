/*
function grecaptchaOnLoad() {
    window.recaptcha = grecaptcha.render('recaptcha', {
        'sitekey': '6LcB8gATAAAAAN4SkOZ0o30BvUFq--YsNiPsIuWp',
    });
}
*/

(function() {
    function setEditorValue(editor, text) {
        editor.setValue(text);
    }

    function getPersistedTextKey() {
        return window.location.href + '::' + 'persistedText';
    }

    function getPersistedLanguageKey() {
        return window.location.href + '::' + 'persistedLanguage';
    }

    function onChange(editor) {
    }

    function setCommonEditorOptions() {
        window.editor.setOption("lineNumbers", true);
        window.editor.setOption("lineWrapping", true);
        window.editor.setOption("theme", "solarized");
        setupEditorSpaces();
    }

    function createEditor(editorElement) {
        window.editor = CodeMirror.fromTextArea(editorElement, {});
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
            setEditorValue(window.editor, text);
        }
    }
    function setupCPP(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "text/x-c++src");
        if (text) {
            setEditorValue(window.editor, text);
        }
    }
    function setupJava(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "text/x-java");
        if (text) {
            setEditorValue(window.editor, text);
        }
    }
    function setupJavaScript(editorElement, text) {
        window.editor.setOption("tabSize", 2);
        window.editor.setOption("indentUnit", 2);
        setCommonEditorOptions();
        window.editor.setOption("mode", "javascript");
        if (text) {
            setEditorValue(window.editor, text);
        }
    }
    function setupPython(editorElement, text) {
        window.editor.setOption("tabSize", 4);
        window.editor.setOption("indentUnit", 4);
        setCommonEditorOptions();
        window.editor.setOption("mode", "python");
        if (text) {
            setEditorValue(window.editor, text);
        }
    }
    function setupRuby(editorElement, text) {
        window.editor.setOption("tabSize", 2);
        window.editor.setOption("indentUnit", 2);
        setCommonEditorOptions();
        window.editor.setOption("mode", "ruby");
        if (text) {
            setEditorValue(window.editor, text);
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
            case 'javascript':
                setupJavaScript(editorElement, text);
                break;
            case 'python':
                setupPython(editorElement, text);
                break;
            case 'ruby':
                setupRuby(editorElement, text);
                break;
        }
    }

    function setCodeButtonsCallbacks(rootUrl) {
        $(".clear-output-button").click(function() {
            $("#output").text("");
        });
        $(".clear-code-button").click(function() {
            setEditorValue(window.editor, "");
            window.editor.clearHistory();
        });
        $(".submit-button").click(function() {
            var l = Ladda.create(document.querySelector('.submit-button'));
            l.start();
            l.isLoading();

            data = {
                'code': window.editor.getValue(),
                //'recaptcha': grecaptcha.getResponse(window.recaptcha),
            };
            var problem = $(".problem-select").val();
            var language = $(".language-select").val();
            var url = rootUrl + "/evaluator/evaluate/" + problem + "/" + language;
            $.ajax({
                type: "POST",
                url: url,
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
    }

    function setupLanguageSelect(editorElement, rootUrl) {
        $(".language-select").change(function() {
            var text = window.editor.getValue();
            language = $(".language-select").val();
            onLanguageSelected(editorElement, language, text);
            refreshProblemSelect(rootUrl).done(function() {
                updateDescriptionAndCode(rootUrl);
            });
        });

        // triggering chosen:updated doesn't trigger the callback function,
        // so we call it ourselves.
        var language = 'python';
        $(".language-select").val(language);
        onLanguageSelected(editorElement, language, window.editor.getValue());
    }

    function refreshProblemSelect(rootUrl) {
        var problemSelect = $(".problem-select");
        problemSelect.prop("disabled", true);
        return $.get(rootUrl + "/evaluator/get_problem_summaries")
         .done(function(problems) {
            problems = _.sortBy(problems, function(problem) {
                return problem.title;
            });
            var language = $(".language-select").val();
            matchingProblems = _.filter(problems, function(problem) {
                return _.includes(problem.supported_languages, language);
            });
            problemSelect.empty();
            if (_.size(matchingProblems) !== 0) {
                _.each(matchingProblems, function(problem) {
                    var option = $("<option></option>")
                                 .attr("value", problem.id)
                                 .text(problem.title);
                    problemSelect.append(option);
                });
            } else {
                var option = $("<option></option>")
                             .attr("value", "")
                             .text("No problems found");
                problemSelect.append(option);
                $("#description").html("<div></div>");
            }
            problemSelect.prop("disabled", false);
         });
    }

    function setupProblemSelect(rootUrl) {
        refreshProblemSelect(rootUrl).done(function() {
            updateDescriptionAndCode(rootUrl);
        });
        $(".problem-select").change(function() {
            updateDescriptionAndCode(rootUrl);
        });
    }

    function updateDescriptionAndCode(rootUrl) {
        problem = $(".problem-select").val();
        language = $(".language-select").val();
        url = rootUrl + "/evaluator/get_problem_details/" + problem + "/" + language;
        return $.get(url).done(function(problem) {
            setEditorValue(window.editor, problem.initial_code[language].code);
            description = marked(problem.description[language].markdown, {
                sanitize: true,
                smartypants: true
            });
            $("#description").html(description);
        });
    }

    //$(window).load(grecaptchaOnLoad);

    $(function() {
        var editorElement = document.getElementById("editor");
        var rootUrl = "http://www.runsomecode.com";
        createEditor(editorElement);
        setCodeButtonsCallbacks(rootUrl);
        setupLanguageSelect(editorElement, rootUrl);
        setupProblemSelect(rootUrl);

        $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
          console.log("refresh");
          window.editor.refresh();
        });
    });
}());
