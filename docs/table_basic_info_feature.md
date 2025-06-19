# Table Basic Info 功能实现

## 概述

该功能为pgweb项目新增了表基本信息展示页面，当用户点击数据库表时，会默认显示表的基本信息而不是表数据行。

## 功能特性

### 后端API
- **接口地址**: `GET /api/db/table_basic_info/{oid}`
- **参数**: `oid` - 表的OID（对象标识符）
- **返回**: TableBasicInfo结构体，包含表的基本信息

### 前端界面
- **新增tab页**: "Basic Info"，作为默认选中的tab
- **显示信息**: 表名、Schema、拥有者、表类型、表空间、注释、分区键、分布键
- **UI设计**: 采用表格形式展示，具有良好的视觉效果

## 实现细节

### 后端实现

#### API层 (`pkg/api/object_info.go`)
```go
func GetTableBasicInfo(c *gin.Context) {
    oidStr := c.Params.ByName("oid")
    oid, err := strconv.Atoi(oidStr)
    
    basicInfo, err := dbClient.TableBasicInfo(oid)
    successResponse(c, basicInfo)
}
```

#### 客户端层 (`pkg/client/client.go`)
```go
func (client *Client) TableBasicInfo(oid int) (*meta.TableBasicInfo, error) {
    dataSourceType := client.GetDataSourceType()
    version := client.ServerVersion()
    metaQueryer := NewDBMetaQueryer(dataSourceType, version)
    
    basicInfo, err := metaQueryer.TableBasicInfo(client.db, oid)
    return &basicInfo, nil
}
```

#### 数据结构 (`pkg/structs/meta/meta.go`)
```go
type TableBasicInfo struct {
    Name           string `json:"name"`
    Schema         string `json:"schema"`
    Comment        string `json:"comment"`
    Table          string `json:"table"`
    TableType      string `json:"table_type"`
    PartitionKeys  string `json:"partition_keys"`
    DistributeKeys string `json:"distribute_keys"`
    TableSpace     string `json:"table_space"`
    Owner          string `json:"owner"`
}
```

### 前端实现

#### HTML结构修改 (`static/index.html`)
```html
<ul>
  <li id="table_basic_info" class="selected">Basic Info</li>
  <li id="table_content">Rows</li>
  <li id="table_structure">Structure</li>
  <!-- 其他tabs -->
</ul>
```

#### JavaScript实现 (`static/js/app.js`)
```javascript
function showTableBasicInfo() {
    var currentObj = getCurrentObject();
    var selectedElement = $("#objects li.active");
    var oid = selectedElement.data("oid");
    
    getTableBasicInfo(oid, function(data) {
        var basicInfoHtml = `
            <div class="table-basic-info">
                <h3>Table Basic Information</h3>
                <table class="table table-bordered">
                    <tbody>
                        <tr><td><strong>Table Name:</strong></td><td>${data.name}</td></tr>
                        <tr><td><strong>Schema:</strong></td><td>${data.schema}</td></tr>
                        <!-- 其他信息行 -->
                    </tbody>
                </table>
            </div>
        `;
        
        $("#results_view").html(basicInfoHtml).show();
    });
}
```

#### CSS样式 (`static/css/app.css`)
```css
.table-basic-info {
    padding: 20px;
    max-width: 800px;
    margin: 0 auto;
}

.table-basic-info h3 {
    color: #333;
    margin-bottom: 20px;
    padding-bottom: 10px;
    border-bottom: 2px solid #79589f;
}

.table-basic-info table td:first-child {
    background-color: #f8f9fa;
    font-weight: 500;
    width: 200px;
}
```

## 用户体验改进

### 默认行为变更
- **之前**: 点击表格 → 跳转到Rows页面查看表数据
- **现在**: 点击表格 → 跳转到Basic Info页面查看表基本信息
- **优势**: 用户可以快速了解表的元数据，而不需要加载大量数据

### 信息展示
显示的信息包括：
- **表名**: 表的完整名称
- **Schema**: 表所属的模式
- **拥有者**: 表的所有者
- **表类型**: 表的类型（如regular table、partitioned table等）
- **表空间**: 表存储的表空间
- **注释**: 表的描述信息
- **分区键**: 分区表的分区键信息
- **分布键**: 分布式表的分布键信息

## 数据源支持

该功能支持以下数据源：
- **PostgreSQL**: 标准PostgreSQL数据库
- **PanweiAP**: 潘微分析平台数据库

不同数据源使用不同的SQL查询来获取表基本信息，通过`DBMetaQueryer`接口实现了统一的访问方式。

## 使用方法

1. 启动pgweb应用
2. 连接到数据库
3. 在左侧面板中点击任意表名
4. 系统自动显示"Basic Info"页面，展示表的基本信息
5. 如需查看表数据，点击"Rows"标签页

## 技术亮点

- **模块化设计**: 后端使用了清晰的分层架构
- **类型安全**: 使用强类型结构体定义数据格式
- **错误处理**: 完善的错误处理机制
- **响应式UI**: 友好的用户界面设计
- **多数据源**: 支持不同类型的数据库 