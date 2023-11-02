var TourName = '';
var Solutions = [];
var Teams = {};
var DayStart = "";
var DayEnd = "";

$( document ).ready(function() {
    $('.select2').select2();
    $('#WishTable').find('.w-from, .w-to').inputmask({alias: "datetime",inputFormat: "HH:MM"});

    drawFields();
    $('#StadID').on('change', drawFields);

    if ( window.location.pathname == '/' ){
        checkStat(true);
    }
    
    $('#GO').on('click', function(){
        var $thisBtn = $(this);
        if ($thisBtn.text()=='Стоп') {
            $thisBtn.attr('disabled', true);
            $.ajax({
                url: '/search-stop',
                type: "POST",
                dataType: "json",
                success: function(data) {
                    $thisBtn.text('Пуск');
                    $thisBtn.removeAttr('disabled');
                }
            });

            return
        }

        var tourName = $('#TourName').val();
        var stadID = parseInt($('#StadID').val());
        
        var fields = [];
        $('#FieldsTable tbody tr').each(function(index){
            var $tr = $(this);
            var format = parseInt($tr.find('.f-format').val());
            if (format < 3 || format > 7) {
                alert('Формат поля должен быть от 3 до 7');
                return;
            }
            fields.push({
                format: format,
                from: $tr.find('.f-from').val(),
                to: $tr.find('.f-to').val(),
                dur: parseInt($tr.find('.f-dur').val()),
            });
        });

        var teams = [];
        $('#TeamsSelects option:selected').each(function(index){
            var teamId = $(this).val();
            teams.push(parseInt(teamId));
        });

        var wishes = [];
        $('#WishTable tbody tr').each(function(index){
            var $tr = $(this);
            wishes.push({
                team_id: parseInt($tr.data('team-id')),
                from: $tr.find('input.w-from').val(),
                to: $tr.find('input.w-to').val()
            });
        });

        var games = [];
        $('#GamesTable tbody tr').each(function(index){
            var $tr = $(this);
            if ($tr.find('input[type="checkbox"]:checked').length == 0) {
                games.push({
                    team_id_1: parseInt($tr.data('team-id-1')),
                    team_id_2: parseInt($tr.data('team-id-2')),
                });
            }
        });

        var data = {
            tour_name: tourName,
            stadium_id: stadID,
            fields: fields,
            teams: teams,
            wishes: wishes,
            games: games
        }

        $('#GO').attr('disabled', true);

        $.ajax({
            url: '/search-start',
            type: "POST",
            data: JSON.stringify(data),
            dataType: "json",
            success: function(data) {
                $('#SolCnt').text(""+data.solutions.length);
                $('#AttCnt').text(data.attempts);
                TourName = data.tour_name;
                Teams = data.teams;
                DayStart = data.day_start.substring(11, 16);
                DayEnd = data.day_end.substring(11, 16);
                loadSolutions(data.solutions);
                $('#GO').removeAttr('disabled');
                $('#GO').text('Стоп');
                $('#Results').css('visibility', 'visible');
            },
            error: function(err){
                console.log('error', err);
            }
        });
    });

    $('#LoadSolutions').on('click', function(){
        $('#LoadSolutions').attr('disabled', true);
        $.ajax({
            url: "/get-solutions",
            type: 'GET',
            dataType: 'json',
            success: function(data) {
                loadSolutions(data.solutions);
                $('#LoadSolutions').removeAttr('disabled', true);
            }
        });
    });

    $('.x-wish-btn').on('click', function(){
        $(this).closest('tr').remove();
    });

    $(document).on('click','.x-field-btn',function(){
        var $tr = $(this).closest('tr');
        var $tbody = $tr.closest('tbody');
        $tr.remove();
        recountFieldIdx($tbody);
    });
    
    $(document).on('click','.solution-item',function(){
        var solutionId = $(this).data('solution-id');
        var $area = $('#SolutionDetailsArea');
        $area.html('');

        $area.append('<h4>'+TourName+', Решение №'+(solutionId+1)+'<a href="/download-solution?hash='+Solutions[solutionId].hash+'" target="blank" class="btn btn-sm btn-primary pull-right"><i class="fa fa-file-excel-o" aria-hidden="true"></i> Скачать</a>'+'<h4>');

        // add tables with bars

        var gamesByTeam = {};
        for (var field in Solutions[solutionId].games) {
            for (var i=0;i<Solutions[solutionId].games[field].length;i++) {
                if (! gamesByTeam[Solutions[solutionId].games[field][i].team_id_1]) {
                    gamesByTeam[Solutions[solutionId].games[field][i].team_id_1] = [];
                }
                if (! gamesByTeam[Solutions[solutionId].games[field][i].team_id_2]) {
                    gamesByTeam[Solutions[solutionId].games[field][i].team_id_2] = [];
                }
                var el = {
                    start: Solutions[solutionId].games[field][i].start.substring(11, 16),
                    end : Solutions[solutionId].games[field][i].end.substring(11, 16)
                };
                gamesByTeam[Solutions[solutionId].games[field][i].team_id_1].push(el);
                gamesByTeam[Solutions[solutionId].games[field][i].team_id_2].push(el);
            }
        }

        $area.append(
            '<table class="tl-table">'+
                timelineTr(gamesByTeam)+
            '</table>'
        );

        for (var field in Solutions[solutionId].games) {
            var $el = fieldTable(field, Solutions[solutionId].games[field], Teams)
            $area.append($el);
        }
    });

    $('.del-btn').on('click', function(){
        var tag = $(this).data('tag');
        var id = $(this).data('id');

        $.ajax({
            type: 'POST',
            url: '/del-entity?tag='+tag+'&id='+id,
            dataType: "json",
            success: function(data) {
                if (data.result) {
                    location.reload();
                    return
                }
                alert(data.error);
            } 
        });
    });

    $('#ModalForm').modal({show:false});

    $('.edit-btn').on('click', function(){
        var $btn = $(this);
        var tag = $btn.data('tag');
        var id = $btn.data('id');
        showForm($btn, tag, id);
    });

    $('#ModalSave').on('click', function(){
        var data = {};
        $('#ModalForm').find('input, select').each(function(){
            var $el = $(this);
            var name = $el.attr('name');
            var val = $el.val();
            if ($el.attr('type')=='checkbox') {
                val = $el.is(":checked") ? "1":"0";
            }
            data[name] = val;
        });

        console.log(data);
        $.ajax({
            type: 'POST',
            url: '/save-entity?tag='+data.tag+'&id='+data.id,
            data: JSON.stringify(data),
            dataType: "json",
            success: function(data) {
                if (data.result) {
                    location.reload();
                    return
                }
                alert(data.error);
            } 
        });
    });

    $('#AddEntity').on('click', function(){
        var $btn = $(this);
        var tag = $btn.data('tag');
        showForm($btn, tag, -1)
    });

    $('#AddField').on('click', function(){
        var fOpts = getFieldOptions();
        $('#FieldsTable tbody').append(getFieldTr(fOpts.format, fOpts.from, fOpts.to, fOpts.dur));
        recountFieldIdx();
    });

    $('.data-table-powered').DataTable({
        responsive: true,
        order: [[1, 'asc']],
        "columnDefs": [ {
            "targets"  : 'no-sort',
            "orderable": false,
        }],
        "pageLength": 500,
        lengthMenu: [
            [10, 25, 50, 100, 500],
            [10, 25, 50, 100, 500],
        ],
        "language": {
            "sProcessing":    "Обработка...",
            "sLengthMenu":    "Показывать _MENU_ на странице",
            "sZeroRecords":   "Пустой результат поиска",
            "sEmptyTable":    "Нет данных",
            "sInfo":          "Показано с _START_ по _END_ из _TOTAL_",
            "sInfoEmpty":     "Показано 0 шт",
            "sInfoFiltered":  "(отфильтровано из _MAX_ шт)",
            "sInfoPostFix":   "",
            "sSearch":        "Искать:",
            "sUrl":           "",
            "sInfoThousands":  ",",
            "sLoadingRecords": "Загрузка...",
            "oPaginate": {
                "sFirst":    "Перв.",
                "sLast":     "Посл.",
                "sNext":     "След.",
                "sPrevious": "Пред."
            },
        }
    });

    recountTeamsSelected();
    $('#TeamsSelects select').on("change", recountTeamsSelected);
});

function recountTeamsSelected() {
    $("#TeamsCnt").html(optsSelectedCnt($("#TeamsSelects")));
    $(".teams-div-selected").each(function(){
        var $that = $(this);
        $that.html(optsSelectedCnt($that.closest('h6').next()));
    });
}
function optsSelectedCnt($el) {
    return $el.find('option:selected').length
}

function timelineTr(gamesByTeam) {
    var str = '';
    for (var teamId in gamesByTeam) {
        str += '<tr>';
        str += '<td style="width:20%">'+Teams[teamId]+'</td>';
        str += '<td>';
        str += '<div class="timeline">';
        for (var i=0; i<gamesByTeam[teamId].length;i++) {
            var start = gamesByTeam[teamId][i].start;
            var end = gamesByTeam[teamId][i].end;
            var iv = getInterval(start, end);
            str += '<div class="timeline-interval" title="'+start+' - '+end+'" style="width:'+iv.width+'%;left:'+iv.left+'%;"></div>'
        }
        str += '</div>';
        str += '</td>';
        str += '</tr>';
    }
    return str;
}
function getInterval(start, end) {
    var dStart = calcHmVal(DayStart);
    var dEnd = calcHmVal(DayEnd);
    var startVal = calcHmVal(start);
    var endVal = calcHmVal(end);
    //console.log(dStart, dEnd, startVal, endVal);
    return {
        left: Math.round((startVal - dStart) / (dEnd-dStart) * 100),
        width: Math.round((endVal-startVal)/(dEnd-dStart) * 100)
    };
}

function calcHmVal(hm) {
    var arr = hm.split(':');
    //console.log(hm, arr)
    return parseInt(arr[0])*60+parseInt(arr[1]);
}

function checkStat(firstCall) {
    $.ajax({
        url: "/status"+(firstCall?'?with-data=1':''),
        type: 'GET',
        dataType: 'json',
        timeout: 2000,
        success: function(data) {
            if (data.status == "process") {
                $('#SolCnt').text(""+data.solutions_cnt);
                $('#AttCnt').text(data.attempts);
                $('#GO').text('Стоп');
                $('#Results').css('visibility', 'visible');
                if (data.tour_name) {
                    TourName = data.tour_name;
                    Teams = data.teams;
                    DayStart = data.day_start.substring(11, 16);
                    DayEnd = data.day_end.substring(11, 16);
                }
            } else {
                $('#GO').text('Пуск').removeAttr('disabled');
                $('#Results').css('visibility', 'hidden');
            }
        }
    }).done(function(){
        setTimeout(checkStat, 2000);
    }).fail(function(){
        setTimeout(checkStat, 2000);
    });
}

function loadSolutions(solutions) {
    var $ul = $('#SolutionsList ul');
    $ul.find('li').remove();
    $('#SolutionDetailsArea').html('');
    Solutions = solutions;
    for (var i = 0; i < solutions.length; i++) {
        $ul.append(
            '<li data-solution-id="'+i+'" class="solution-item list-group-item d-flex justify-content-between align-items-center">'+
                'Решение №'+(i+1)+
                '<span class="badge badge-info badge-pill">'+solutions[i].sum+'</span>'+
            ' </li>'
        );
    }
}

function getFieldOptions() {
    var $opt = $('#StadID').find(':selected');
    var fields = $opt.data('fields');
    var format = $opt.data('format');
    var from = $opt.data('from');
    var to = $opt.data('to');
    var dur = $opt.data('game-dur');

    return {
        fields: fields,
        format: format,
        from: from,
        to: to,
        dur: dur,
    }
}

function getFieldTr(format, from, to, dur) {
    return $(
        '<tr>'+
            '<td><input type="text" class="form-control form-control-sm" value="0" disabled></td>'+
            '<td><input type="text" class="form-control form-control-sm f-format" value="'+format+'"></td>'+
            '<td><input type="text" class="form-control form-control-sm f-from" value="'+from+'" disabled></td>'+
            '<td><input type="text" class="form-control form-control-sm f-to" value="'+to+'"></td>'+
            '<td><input type="text" class="form-control form-control-sm f-dur" value="'+dur+'" disabled></td>'+
            '<td><button type="button" class="btn btn-sm btn-warning x-field-btn">✕</button></td>'+
        '</tr>'
    )
}

function recountFieldIdx() {
    var $tbody = $('#FieldsTable tbody')
    $tbody.find('tr').each(function(index){
        $(this).find('td:first-child input').val(index+1);
    });
}

function drawFields() {
    var fOpts = getFieldOptions();
    $('#FieldsTable tbody tr').remove();
    for (var i=0;i<fOpts.fields; i++) {
        $('#FieldsTable tbody').append(getFieldTr(fOpts.format, fOpts.from, fOpts.to, fOpts.dur));
    }
    recountFieldIdx();
    $('#FieldsTable').find('.f-from, .f-to').inputmask({alias: "datetime",inputFormat: "HH:MM"});
    $('#FieldsTable').find('.f-format, .f-dur').inputmask({ regex: "^[0-9]{1,3}$" });
}

function fieldTable(fieldName, games, teams) {
    var gamesHtml = '';

    for (var i=0; i<games.length; i++) {
        gamesHtml +=
        '<tr>'+
            '<td>'+teams[games[i].team_id_1]+'</td>'+
            '<td>'+teams[games[i].team_id_2]+'</td>'+
            '<td>'+games[i].start.substring(11, 16)+'-'+games[i].end.substring(11, 16)+'</td>'+
        '</tr>'
    }

    return $('<h5>'+fieldName+'</h5>'+
        '<table id="WishTable" class="table table-sm">'+
            '<thead class="thead-light">'+
                '<tr>'+
                    '<th scope="col" style="width:40%">Команда1</th>'+
                    '<th scope="col">Команда2</th>'+
                    '<th scope="col" style="width:100px">Время</th>'+
                '</tr>'+
            '</thead>'+
            '<tbody>'+
                gamesHtml+
            '<tbody>'+
        '</table>'
    );
}

var formTemplates = {
    stadium: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="stadName">Название</label>
            <input name="name" type="text" class="form-control form-control-sm" id="stadName">
        </div>
        <div class="form-group">
            <label for="fieldsCount">Количество полей</label>
            <input name="fields" type="text" class="form-control form-control-sm" id="fieldsCount">
        </div>
        <div class="form-group">
            <label for="fieldsFormat">Формат</label>
            <input name="format" type="text" class="form-control form-control-sm" id="fieldsFormat">
        </div>
        <div class="form-group">
            <label for="fromTime">Работает с</label>
            <input name="time_from" type="text" class="form-control form-control-sm" id="fromTime">
        </div>
        <div class="form-group">
            <label for="toTime">Работает по</label>
            <input name="time_to" type="text" class="form-control form-control-sm" id="toTime">
        </div>
        <div class="form-group">
            <label for="gameDur">Продолжительность игры (мин.)</label>
            <input name="game_dur" type="text" class="form-control form-control-sm" id="gameDur">
        </div>
    `,
    division: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="stadName">Название</label>
            <input name="name" type="text" class="form-control form-control-sm" id="stadName">
        </div>
        <div class="form-group">
            <label for="fieldsFormat">Формат</label>
            <input name="format" type="text" class="form-control form-control-sm" id="fieldsFormat">
        </div>
    `,
    coach: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="stadName">Имя</label>
            <input name="name" type="text" class="form-control form-control-sm" id="stadName">
        </div>
    `,
    team: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="stadName">Название</label>
            <input name="name" type="text" class="form-control form-control-sm" id="stadName">
        </div>
        <div class="form-group">
            <label for="divisionId">Дивизион</label>
            <select name="division_id" class="form-control form-control-sm select2" id="divisionId" style="width:100%"></select>
        </div>
        <div class="form-group">
            <label for="coachId">Тренер</label>
            <select name="coach_id" class="form-control form-control-sm select2" id="coachId" style="width:100%"></select>
        </div>
    `,
    wish: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="teamId">Название</label>
            <select name="team_id" class="form-control form-control-sm select2" id="teamId" style="width:100%"></select>
        </div>
        <div class="form-group">
            <label for="timeFrom">С</label>
            <input name="time_from" type="text" class="form-control form-control-sm" id="timeFrom">
        </div>
        <div class="form-group">
            <label for="timeTo">По</label>
            <input name="time_to" type="text" class="form-control form-control-sm" id="timeTo">
        </div>
    `,
    games: `
        <div class="form-group">
            <input type="hidden" name="id" id="stadId">
            <input type="hidden" name="tag" id="stadTag">
            <label for="tourName">Тур</label>
            <input name="tour" type="text" class="form-control form-control-sm" id="tourName">
        </div>
        <div class="form-group">
            <label for="teamID1">Команда1</label>
            <select name="team_id_1" class="form-control form-control-sm select2" id="teamID1" style="width:100%"></select>
        </div>
        <div class="form-group">
            <label for="teamID2">Команда1</label>
            <select name="team_id_2" class="form-control form-control-sm select2" id="teamID2" style="width:100%"></select>
        </div>
        <div class="form-check">
            <input name="can_rematch" value="1" type="checkbox" class="form-check-input" id="canRematch">
            <label class="form-check-label" for="canRematch">Переигровка возможна</label>
        </div>
    `
};

function showForm($btn, tag, id) {
    var $html;
    switch (tag) {
        case 'stadium':
            $html = $(formTemplates.stadium);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#stadName').val($tds.eq(1).html());
                $html.find('#fieldsCount').val($tds.eq(2).html());
                $html.find('#fieldsFormat').val($tds.eq(3).html());
                $html.find('#fromTime').val($tds.eq(4).html());
                $html.find('#toTime').val($tds.eq(5).html());
                $html.find('#gameDur').val(parseInt($tds.eq(6).html()));
            }
            $html.find('#fieldsCount, #fieldsFormat').inputmask("9");
            $html.find('#gameDur').inputmask({ regex: "^[0-9]{1,3}$" });
            $html.find('#fromTime, #toTime').inputmask({alias: "datetime",inputFormat: "HH:MM"});
            break;
        case 'division':
            $html = $(formTemplates.division);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#stadName').val($tds.eq(1).html());
                $html.find('#fieldsFormat').val($tds.eq(2).html());
            }
            $html.find('#fieldsFormat').inputmask("9");
            break;
        case 'coach':
            $html = $(formTemplates.coach);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#stadName').val($tds.eq(1).html());
            }
            break;
        case 'team':
            $html = $(formTemplates.team);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            var $divEl = $html.find('#divisionId');
            var $coachEl = $html.find('#coachId');
            for (k in Divisions) {
                $divEl.append('<option value="'+k+'">'+Divisions[k]+'</option>');
            }
            for (k in Coaches) {
                $coachEl.append('<option value="'+k+'">'+Coaches[k]+'</option>');
            }
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#stadName').val($tds.eq(1).html());
                $divEl.val($btn.data('div-id'));
                $coachEl.val($btn.data('coach-id'));
            }
            $html.find('.select2').select2();
            break;
        case 'wish':
            $html = $(formTemplates.wish);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            var $teamIdEl = $html.find('#teamId');
            for (k in Teams) {
                $teamIdEl.append('<option value="'+k+'">'+Teams[k]+'</option>');
            }
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#timeFrom').val($tds.eq(1).html());
                $html.find('#timeTo').val($tds.eq(2).html());
                $teamIdEl.val($btn.data('team-id'));
            }
            $html.find('#timeFrom').inputmask({alias: "datetime",inputFormat: "HH:MM"});
            $html.find('#timeTo').inputmask({alias: "datetime",inputFormat: "HH:MM"});
            $html.find('.select2').select2();
            break;
        case 'game':
            $html = $(formTemplates.games);
            $html.find('#stadId').val(id);
            $html.find('#stadTag').val(tag);
            var $teamIdEl1 = $html.find('#teamID1');
            var $teamIdEl2 = $html.find('#teamID2');
            console.log("len", Teams.length)
            for (k in Teams) {
                $teamIdEl1.append('<option value="'+k+'">'+Teams[k]+'</option>');
                $teamIdEl2.append('<option value="'+k+'">'+Teams[k]+'</option>');
            }
            if (id != -1) {
                var $tds = $btn.closest('tr').find('td');
                $html.find('#tourName').val($tds.eq(0).html());
                var $canRematchEl = $html.find('#canRematch');
                $teamIdEl1.val($btn.data('team-id-1'));
                $teamIdEl2.val($btn.data('team-id-2'));
                if ($tds.eq(3).html() == '1') {
                    $canRematchEl.attr('checked','checked');
                }
            }
            $html.find('.select2').select2();
            break;
        default:
            return;
    }
    // formTemplates.stadium

    $('#ModalForm').find('.modal-body').html($html);
    $('#ModalForm').modal('toggle');
}
