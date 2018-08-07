package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_weixin_module/library"
	"reflect"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	//经过查阅多方资料，找到的答案是指定具体连接mysql的方式有三种不同代码：
	//①
	//db, err := sql.Open("mysql", "user:password@unix(/tmp/mysql.sock)/test")
	//②
	//db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/test")   //指定IP和端口
	//③
	//db, err := sql.Open("mysql", "user:password@/test")  //默认方式

	//"root:guotingyu0324@tcp(192.168.81.129:3306)/webcms?charset=utf8"
	dbSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", beego.AppConfig.String("mysqluser"), beego.AppConfig.String("mysqlpass"), beego.AppConfig.String("mysqlurls"), beego.AppConfig.String("mysqldb"), beego.AppConfig.String("mysqlcharset"))
	orm.RegisterDataBase("default", "mysql", dbSource)
}

type Base struct {
	TableName        string
	ReorganizeFields map[string]string
}

// GetInfoById
func (t *Base) reorganize(data map[string]interface{}) map[string]interface{} {
	for field, handler := range t.ReorganizeFields {
		if val, ok := data[field]; ok {
			if handler == "[]string" {
				var slice1 []string
				err := json.Unmarshal([]byte(library.Strval(val)), &slice1)
				if err != nil {
					fmt.Println("reorganize error:", field, val, handler)
					panic(err)
				}
				data[field] = slice1
			} else if handler == "map[string]interface{}" {
				var map1 map[string]interface{}
				err := json.Unmarshal([]byte(library.Strval(val)), &map1)
				if err != nil {
					fmt.Println("reorganize error:", field, val, handler)
					panic(err)
				}
				data[field] = map1
			}
		}
	}
	return data
}

// GetInfoById
func (t *Base) GetInfoById(id string) (maps map[string]interface{}) {
	query := make(map[string]interface{})
	query["_id"] = id
	info := t.FindOne([]map[string]interface{}{query})
	return info
}

// count
func (t *Base) Count(query []map[string]interface{}) int64 {
	sqlAndConditions := t.getSqlAndConditions4Count(query)
	//fmt.Println(sqlAndConditions)
	sql, args := t.getArgs(sqlAndConditions)
	num, data := t.executeQuery(sql, args, "query")
	//fmt.Println("num:", num, "data:", data)
	info := make(map[string]interface{})
	var countNum int64
	if num > 0 {
		for _, row := range data {
			info = (map[string]interface{})(row)
			countNum := library.Intval(info["num"])
			//fmt.Println("countNum:", countNum, "data_num:", info["num"])
			return countNum
		}
	}
	return countNum
}

// findOne
func (t *Base) FindOne(query []map[string]interface{}) (maps map[string]interface{}) {
	sqlAndConditions := t.getSqlAndConditions4FindOne(query)
	//fmt.Println(sqlAndConditions)
	sql, args := t.getArgs(sqlAndConditions)
	num, data := t.executeQuery(sql, args, "query")
	info := make(map[string]interface{})
	if num > 0 {
		for _, row := range data {
			info = t.reorganize((map[string]interface{})(row))
		}
	}
	return info
}

// findAll
func (t *Base) FindAll(query []map[string]interface{}, sort []map[string]interface{}, fields map[string]interface{}) (maps []map[string]interface{}) {
	sqlAndConditions := t.getSqlAndConditions4FindAll(query, sort, fields)
	//fmt.Println(sqlAndConditions)
	sql, args := t.getArgs(sqlAndConditions)
	fmt.Println(sql)
	num, data := t.executeQuery(sql, args, "query")
	list := make([]map[string]interface{}, 0)
	if num > 0 {
		for _, row := range data {
			info := t.reorganize((map[string]interface{})(row))
			list = append(list, info)
		}
	}
	return list

}

// find
func (t *Base) Find(query []map[string]interface{}, sort []map[string]interface{}, skip int64, limit int64, fields map[string]interface{}) map[string]interface{} {
	fmt.Println("find query:", query)
	total := t.Count(query)
	fmt.Println("total:", total)
	sqlAndConditions := t.getSqlAndConditions4Find(query, sort, skip, limit, fields)
	sql, args := t.getArgs(sqlAndConditions)
	fmt.Println("find query:", sqlAndConditions, sql, args)
	num, data := t.executeQuery(sql, args, "query")
	list := make([]map[string]interface{}, 0)
	if num > 0 {
		for _, row := range data {
			info := t.reorganize((map[string]interface{})(row))
			list = append(list, info)
		}
	}
	ret := make(map[string]interface{}, 0)
	ret["total"] = total
	ret["datas"] = list
	//fmt.Println("Find:", ret)
	return ret
}

//执行insert操作
func (t *Base) Insert(contents map[string]interface{}) (maps map[string]interface{}) {
	sqlAndConditions := t.getSqlAndConditions4Insert(contents)
	sql, args := t.getArgs(sqlAndConditions)
	fmt.Println(sqlAndConditions, sql, args)
	num, _ := t.executeQuery(sql, args, "execute")
	if num < 1 {
		panic(errors.New("追击数据的时候发生错误"))
	}
	query := make(map[string]interface{})
	insertFieldValues, _ := sqlAndConditions["insertFieldValues"].(map[string]interface{})
	query["_id"] = insertFieldValues["_id"].(string)
	return t.FindOne([]map[string]interface{}{query})
}

// 执行Update操作
func (t *Base) Update(query []map[string]interface{}, contents map[string]interface{}, options map[string]interface{}) int64 {
	sqlAndConditions := t.getSqlAndConditions4Update(query, contents, options)
	sql, args := t.getArgs(sqlAndConditions)
	//fmt.Println(sqlAndConditions, sql, args)
	num, _ := t.executeQuery(sql, args, "execute")
	return num
}

// 执行delete操作
func (t *Base) Remove(query []map[string]interface{}) int64 {
	sqlAndConditions := t.getSqlAndConditions4Remove(query)
	sql, args := t.getArgs(sqlAndConditions)
	//fmt.Println(sqlAndConditions, sql, args)
	num, _ := t.executeQuery(sql, args, "execute")
	return num
}

// executeQuery
func (t *Base) executeQuery(sql string, args []interface{}, method string) (num int64, maps []orm.Params) {
	// '/:(.*?):/i'
	reg := regexp.MustCompile(`:(.*?):`)
	sql = reg.ReplaceAllString(sql, "?")

	// '/\[(.*?)\]/i'
	reg = regexp.MustCompile(`\[(.*?)\]`)
	sql = reg.ReplaceAllString(sql, "`$1`")

	o := orm.NewOrm()
	//fmt.Println(sql, args)
	if method == "query" {

		if _, ok := args["for_update"]; ok {
			sql = sql + "  FOR UPDATE "
			delete(args, "for_update")
		}

		num, err := o.Raw(sql, args...).Values(&maps)
		if err != nil {
			panic(err)
		}
		return num, maps
	} else {
		//fmt.Println(sql, args)
		sqlResult, err := o.Raw(sql, args...).Exec()
		if err != nil {
			panic(err)
		}
		num, err = sqlResult.RowsAffected()
		if err != nil {
			panic(err)
		}
		return num, nil
	}

}

func (t *Base) getArgs(sqlAndConditions map[string]interface{}) (string, []interface{}) {

	sql, _ := sqlAndConditions["sql"].(string)
	data := make(map[string]interface{})
	if _, ok := sqlAndConditions["insertFieldValues"]; ok {
		insertFieldValues, _ := sqlAndConditions["insertFieldValues"].(map[string]interface{})
		insertValues, _ := insertFieldValues["values"].(map[string]interface{})
		for key, value := range insertValues {
			data[key] = value
		}
	}

	if _, ok := sqlAndConditions["conditions"]; ok {
		conditions, _ := sqlAndConditions["conditions"].(map[string]interface{})
		bindValues, _ := conditions["bind"].(map[string]interface{})
		//fmt.Println(bindValues)
		for key, value := range bindValues {
			data[key] = value
		}
	}

	if _, ok := sqlAndConditions["updateFieldValues"]; ok {
		updateFieldValues, _ := sqlAndConditions["updateFieldValues"].(map[string]interface{})
		updateValues, _ := updateFieldValues["values"].(map[string]interface{})
		for key, value := range updateValues {
			data[key] = value
		}
	}

	reg := regexp.MustCompile(`:(.*?):`)
	fields := reg.FindAllString(sql, -1)
	args := make([]interface{}, 0)
	for _, field := range fields {
		if value, ok := data[strings.Trim(field, ":")]; ok {
			args = append(args, value)
		}
	}
	return sql, args
}

func (t *Base) changeValue4Conditions(value interface{}, field string) interface{} {
	//    if (field == '_id') {
	//        value = $this->getMongoId4Query(value);
	//        // die("_id's value:" . value);
	//        return value;
	//    } else {
	//        if (is_bool(value)) {
	//            value = intval(value);
	//            return value;
	//        }
	//        if (value instanceof \MongoDate) {
	//            value = date('Y-m-d H:i:s', value->sec);
	//            return value;
	//        }
	//        if (value instanceof \MongoRegex) {
	//            // /系统管理员/i->'%Art%'
	//            value = value->__toString();
	//            value = str_ireplace('/i', '%', value);
	//            value = str_ireplace('/^$', '', value);
	//            value = str_ireplace('/', '%', value);
	//            return value;
	//        }
	//    }

	if library.Is_bool(value) {
		return library.Intval(value)
	}
	return value
}

func (t *Base) changeValue4Save(value interface{}) interface{} {

	//    if (value instanceof \MongoDate) {
	//        value = date('Y-m-d H:i:s', value->sec);
	//    } elseif (is_bool(value)) {
	//        value = intval(value);
	//    } elseif (is_array(value)) {
	//        if (! empty(value)) {
	//            value = json_encode(value);
	//        } else {
	//            value = "";
	//        }
	//    } elseif (is_object(value)) {
	//        value = json_encode(value);
	//    }

	targetValue := reflect.ValueOf(value)
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Array:
		return library.Json_encode(value)
	case reflect.Map:
		return library.Json_encode(value)
	case reflect.String:
		return value
	case reflect.Bool:
		if targetValue.Bool() {
			return 1
		} else {
			return 0
		}
	}

	return value
}

func (t *Base) getSqlAndConditions4Count(query []map[string]interface{}) map[string]interface{} {
	conditions := t.getConditions(query, "AND")
	if len(conditions) == 0 {
		conditions["conditions"] = "1=1"
		conditions["bind"] = make(map[string]interface{})
	}
	tableName := t.TableName
	phql := fmt.Sprintf("SELECT COUNT(*) as num FROM [%s] WHERE %s ", tableName, conditions["conditions"])
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	return ret
}

func (t *Base) getSqlAndConditions4FindOne(query []map[string]interface{}) map[string]interface{} {
	conditions := t.getConditions(query, "AND")
	if len(conditions) == 0 {
		conditions["conditions"] = "1=1"
		conditions["bind"] = make(map[string]interface{})
	}
	tableName := t.TableName
	phql := fmt.Sprintf("SELECT * FROM [%s] WHERE %s limit 1", tableName, conditions["conditions"])
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	return ret
}

func (t *Base) getSqlAndConditions4FindAll(query []map[string]interface{}, sort []map[string]interface{}, fields map[string]interface{}) map[string]interface{} {
	conditions := t.getConditions(query, "AND")
	if len(conditions) == 0 {
		conditions["conditions"] = "1=1"
		conditions["bind"] = make(map[string]interface{})
	}
	tableName := t.TableName
	order := t.getSort(sort)
	orderBy := ""
	if _, ok := order["order"]; ok {
		orderBy = fmt.Sprintf("ORDER BY %s", order["order"])
	}
	//conditions = array_merge(conditions, order)
	for field, value := range order {
		conditions[field] = value
	}
	phql := fmt.Sprintf("SELECT * FROM [%s] WHERE %s %s ", tableName, conditions["conditions"], orderBy)
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	return ret
}

func (t *Base) getSqlAndConditions4Find(query []map[string]interface{}, sort []map[string]interface{}, skip int64, limit int64, fields map[string]interface{}) map[string]interface{} {
	conditions := t.getConditions(query, "AND")
	if len(conditions) == 0 {
		conditions["conditions"] = "1=1"
		conditions["bind"] = make(map[string]interface{})
	}
	tableName := t.TableName
	order := t.getSort(sort)
	orderBy := ""
	if _, ok := order["order"]; ok {
		orderBy = fmt.Sprintf("ORDER BY %s", order["order"])
	}
	//conditions = array_merge(conditions, order)
	for field, value := range order {
		conditions[field] = value
	}
	phql := fmt.Sprintf("SELECT * FROM [%s] WHERE %s %s  LIMIT %d OFFSET %d ", tableName, conditions["conditions"], orderBy, limit, skip)
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	return ret
}

func (t *Base) getSqlAndConditions4Insert(contents map[string]interface{}) map[string]interface{} {

	tableName := t.TableName
	insertFieldValues := t.getInsertContents(contents)
	phql := fmt.Sprintf("INSERT INTO [%s](%s) VALUES (%s)", tableName, insertFieldValues["fields"], insertFieldValues["bindFields"])

	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["insertFieldValues"] = insertFieldValues
	return ret
}

func (t *Base) getSqlAndConditions4Update(query []map[string]interface{}, contents map[string]interface{}, options map[string]interface{}) map[string]interface{} {
	if len(query) == 0 {
		panic(errors.New("更新数据的时候请指定条件"))
	}

	conditions := t.getConditions(query, "AND")
	updateFieldValues := t.getUpdateContents(contents)

	tableName := t.TableName
	phql := fmt.Sprintf("UPDATE [%s] SET %s WHERE %s ", tableName, updateFieldValues["fields"], conditions["conditions"])
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	ret["updateFieldValues"] = updateFieldValues
	return ret
}

func (t *Base) getSqlAndConditions4Remove(query []map[string]interface{}) map[string]interface{} {
	if len(query) == 0 {
		//panic(errors.New("删除数据的时候请指定条件"))
		query = make([]map[string]interface{}, 0)
	}

	conditions := t.getConditions(query, "AND")
	if len(conditions) == 0 {
		conditions["conditions"] = "1=1"
		conditions["bind"] = make(map[string]interface{}, 0)
	}
	tableName := t.TableName
	phql := fmt.Sprintf("DELETE FROM [%s] WHERE %s ", tableName, conditions["conditions"])
	ret := make(map[string]interface{})
	ret["sql"] = phql
	ret["conditions"] = conditions
	return ret
}

func (t *Base) getInsertContents(contents map[string]interface{}) map[string]interface{} {
	var fields []string
	var bindFields []string
	values := make(map[string]interface{})

	if len(contents) == 0 {
		panic(errors.New("字段没有定义"))
	}

	if _, ok := contents["_id"]; !ok {
		contents["_id"] = library.GetMongoId()
	}

	currentTime := library.GetCurrentTime()
	contents["__CREATE_TIME__"] = currentTime
	contents["__MODIFY_TIME__"] = currentTime
	contents["__REMOVED__"] = false

	for field, value := range contents {
		fieldKey := fmt.Sprintf("[%s]", field)
		fields = append(fields, fmt.Sprintf("%s", fieldKey))
		fieldBindKey := fmt.Sprintf("[%s]_1", field)
		bindFields = append(bindFields, fmt.Sprintf(":%s:", fieldBindKey))
		values[fieldBindKey] = t.changeValue4Save(value)
	}
	if len(fields) == 0 {
		panic(errors.New("字段没有定义"))
	} else {
		ret := make(map[string]interface{})
		ret["fields"] = strings.Join(fields, ",")
		ret["bindFields"] = strings.Join(bindFields, ",")
		ret["values"] = values
		ret["_id"] = contents["_id"]
		return ret
	}
}

func (t *Base) getUpdateContents(contents map[string]interface{}) map[string]interface{} {
	var fields []string
	values := make(map[string]interface{})

	if len(contents) == 0 {
		panic(errors.New("更新字段没有定义"))
	}
	for key, items := range contents {
		switch key {
		case "$exp":
			items2, _ := items.(map[string]interface{})
			if len(items2) > 0 {
				for field, value := range items2 {
					fieldKey := fmt.Sprintf("[%s]", field)
					fields = append(fields, fmt.Sprintf("%s=%s", fieldKey, value))
				}
			}
		case "$set":
			items2, _ := items.(map[string]interface{})
			if len(items2) > 0 {
				for field, value := range items2 {
					fieldKey := fmt.Sprintf("[%s]", field)
					fields = append(fields, fmt.Sprintf("%s=:%s:", fieldKey, field))
					values[field] = t.changeValue4Save(value)
				}
			}
		case "$inc":
			items2, _ := items.(map[string]interface{})
			if len(items2) > 0 {
				for field, value := range items2 {
					fieldKey := fmt.Sprintf("[%s]", field)
					value = t.changeValue4Save(value)
					fields = append(fields, fmt.Sprintf("%s=%s+%d", fieldKey, fieldKey, value))
				}
			}
		default:
			panic(errors.New("更新类别没有定义"))
		}
	}

	if len(fields) == 0 {
		panic(errors.New("更新字段没有定义"))
	} else {
		field := "__MODIFY_TIME__"
		value := library.GetCurrentTime()
		fieldKey := fmt.Sprintf("[%s]", field)
		fields = append(fields, fmt.Sprintf("%s=:%s:", fieldKey, field))
		values[field] = t.changeValue4Save(value)

		ret := make(map[string]interface{})
		ret["fields"] = strings.Join(fields, ",")
		ret["values"] = values
		return ret
	}
}

func (t *Base) getConditions(where []map[string]interface{}, condition_op string) map[string]interface{} {
	unique := library.Uniqid(8)
	var conditions []string
	bind := make(map[string]interface{})
	forUpdate := make(map[string]interface{})

	for _, whereitem := range where {
		for key, item := range whereitem {
			// 如果__FOR_UPDATE__ 存在的话
			if key == "__FOR_UPDATE__" {
				//存在
				forUpdate["for_update"] = item
				delete(whereitem, key)

			} else if key == "__QUERY_OR__" { // 如果__QUERY_OR__ 存在的话
				condition_op = "OR"
				orConditions := item.([][]map[string]interface{})
				delete(whereitem, key)
				for _, condition := range orConditions {
					andConditions := t.getConditions(condition, "AND")
					if len(andConditions) > 1 {
						conditions = append(conditions, library.Strval(andConditions["conditions"]))
						// bind = array_merge(bind, orConditions["bind"])
						for field, value := range andConditions["bind"].(map[string]interface{}) {
							bind[field] = value
						}
					}
				}
			} else if key == "__OR__" {
				// 解决OR查询
				orConditions := t.getConditions(item.([]map[string]interface{}), "OR")
				if len(orConditions) > 1 {
					conditions = append(conditions, library.Strval(orConditions["conditions"]))
					// bind = array_merge(bind, orConditions["bind"])
					for field, value := range orConditions["bind"].(map[string]interface{}) {
						bind[field] = value
					}
				}
			} else {
				fieldKey := "[" + key + "]"
				bindKey := "__" + key + unique + "__"

				if item2, ok1 := item.(map[string]interface{}); ok1 { //reflect.TypeOf(item) == reflect.Array
					//item2 := make(map[string]interface{})
					for op, value := range item2 {
						//value = t.changeValue4Conditions(value, key)
						if op == "$in" {
							value1 := library.Array_values(value)
							if len(value1) > 0 {
								// conditions[] = "{fieldKey} IN ({{bindKey}:array})"
								// bind[bindKey] = array_values(value)
								var bindKey4InArr []string
								for idex, item := range value1 {
									//bindKey4In[] = bindKey + '_' + idex
									tempKey := fmt.Sprintf("%s_%d", bindKey, idex)
									bindKey4InArr = append(bindKey4InArr, tempKey)
									bind[tempKey] = item
								}

								bindKey4In := strings.Join(bindKey4InArr, ":,:")
								//conditions[] = "{fieldKey} IN (:{bindKey4In}:)"
								condtionTmp := fmt.Sprintf("%s IN (:%s:)", fieldKey, bindKey4In)
								conditions = append(conditions, condtionTmp)
								//fmt.Println(condtionTmp)

							} else {
								//conditions[] = "{fieldKey}=:{bindKey}:"
								condtionTmp := fmt.Sprintf("%s=:%s:", fieldKey, bindKey)
								conditions = append(conditions, condtionTmp)
								bind[bindKey] = ""
							}
						}

						if op == "$nin" {
							value1 := library.Array_values(value)
							if len(value1) > 0 {
								// conditions[] = "{fieldKey} NOT IN ({{bindKey}:array})"
								// bind[bindKey] = array_values(value)
								var bindKey4InArr []string
								for idex, item := range value1 {
									//bindKey4In[] = bindKey + '_' + idex
									tempKey := fmt.Sprintf("%s_%d", bindKey, idex)
									bindKey4InArr = append(bindKey4InArr, tempKey)
									bind[tempKey] = item
								}

								bindKey4In := strings.Join(bindKey4InArr, ":,:")
								//conditions[] = "{fieldKey} NOT IN (:{bindKey4In}:)"
								condtionTmp := fmt.Sprintf("%s NOT IN (:%s:)", fieldKey, bindKey4In)
								conditions = append(conditions, condtionTmp)
								//fmt.Println(condtionTmp)

							} else {
								//conditions[] = "{fieldKey}!=:{bindKey}:"
								condtionTmp := fmt.Sprintf("%s!=:%s:", fieldKey, bindKey)
								conditions = append(conditions, condtionTmp)
								bind[bindKey] = ""
							}
						}

						if op == "$ne" {
							//conditions[] = "{fieldKey}!=:{bindKey}:"
							condtionTmp := fmt.Sprintf("%s!=:%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind[bindKey] = value
						}
						if op == "$lt" {
							//conditions[] = "{fieldKey}<:lt_{bindKey}:"
							condtionTmp := fmt.Sprintf("%s<:lt_%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind["lt_"+bindKey] = value
						}
						if op == "$lte" {
							//conditions[] = "{fieldKey}<=:lte_{bindKey}:"
							condtionTmp := fmt.Sprintf("%s<=:lte_%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind["lte_"+bindKey] = value
						}

						if op == "$gt" {
							//conditions[] = "{fieldKey}>:gt_{bindKey}:"
							condtionTmp := fmt.Sprintf("%s>:gt_%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind["gt_"+bindKey] = value
						}
						if op == "$gte" {
							//conditions[] = "{fieldKey}>=:gte_{bindKey}:"
							condtionTmp := fmt.Sprintf("%s>=:gte_%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind["gte_"+bindKey] = value
						}

						if op == "$like" {
							// 解决like查询
							//conditions[] = "{fieldKey} LIKE :like_{bindKey}:"
							condtionTmp := fmt.Sprintf("%s LIKE :like_%s:", fieldKey, bindKey)
							conditions = append(conditions, condtionTmp)
							bind["like_"+bindKey] = value
						}

					}
				} else {
					if ok2 := false; ok2 { //item instanceof \MongoRegex
						//conditions[] = "{fieldKey} LIKE :{bindKey}:"
						condtionTmp := fmt.Sprintf("%s LIKE :%s:", fieldKey, bindKey)
						conditions = append(conditions, condtionTmp)
					} else {
						//conditions[] = "{fieldKey} = :{bindKey}:"
						condtionTmp := fmt.Sprintf("%s = :%s:", fieldKey, bindKey)
						conditions = append(conditions, condtionTmp)
					}
					value := t.changeValue4Conditions(item, key)
					bind[bindKey] = value
				}
			}
		}
	}
	ret := make(map[string]interface{})
	if len(bind) > 0 {
		ret["conditions"] = "(" + strings.Join(conditions, " "+condition_op+" ") + ")"
		ret["bind"] = bind
		for field, value := range forUpdate {
			ret[field] = value
		}
	}
	return ret
}

func (t *Base) getSort(sort []map[string]interface{}) map[string]interface{} {
	var order []string
	for _, item := range sort {
		for key, value := range item {
			if key == "__RANDOM__" {
				// 解决随机查询
				order = append(order, "rand()")
			} else {
				//fieldKey := "[{key}]"
				fieldKey := fmt.Sprintf("[%s]", key)
				value1 := library.Intval(value)
				if value1 > 0 {
					order = append(order, fmt.Sprintf("%s ASC", fieldKey))
				} else {
					order = append(order, fmt.Sprintf("%s DESC", fieldKey))
				}
			}
		}
	}
	ret := make(map[string]interface{})
	if len(order) > 0 {
		ret["order"] = strings.Join(order, ",")
	}
	return ret
}
