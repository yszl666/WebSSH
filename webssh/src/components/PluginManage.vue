<template>
  <el-tab-pane label="插件管理" name="pluginManage">
    <!-- ========================================= -->
    <el-card>
      <el-row>
        <el-table :data="data.plugin_list" style="width: 100%" :show-overflow-tooltip="true">
          <el-table-column fixed sortable prop="name" label="插件名称"></el-table-column>
          <el-table-column sortable prop="title" label="插件标题"></el-table-column>
          <el-table-column sortable prop="description" label="插件描述" width="300"></el-table-column>
          <el-table-column sortable prop="version" label="版本号"></el-table-column>
          <el-table-column sortable prop="author" label="作者"></el-table-column>
          <el-table-column sortable prop="status" label="状态">
            <template #default="scope">
              {{ scope.row.status === 'enabled' ? '启用' : '禁用' }}
          </template>
          </el-table-column>
          <el-table-column sortable prop="created_at" label="创建时间" width="180"></el-table-column>
          <el-table-column fixed="right" label="操作">
            <template #header>
              <el-button type="primary" @click="addPlugin">新增</el-button>
              <el-button type="primary" @click="getPluginList(0, 10000)">刷新</el-button>
            </template>
            <template #default="scope">
              <el-button type="primary" @click="editPlugin(scope.row)">编辑</el-button>
              <el-button type="success" @click="togglePluginStatus(scope.row)">
                {{ scope.row.status === 'enabled' ? '禁用' : '启用' }}
              </el-button>
              <el-button type="info" @click="viewPlugin(scope.row)">预览</el-button>
              <el-button type="primary" @click="viewPlugin(scope.row)">访问</el-button>
              <el-popconfirm confirmButtonText="删除" cancelButtonText="取消" icon="el-icon-info" iconColor="red"
                title="确定删除吗" @confirm="deletePluginById(scope.row.id)">
                <template #reference>
                  <el-button type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-row>
    </el-card>
    <!-- ========================================= -->
    <el-dialog title="插件信息" v-model="data.plugin_dialog_visible" width="60%">
      <el-form label-width="80px">
        <el-form-item label="插件名称" prop="name">
          <el-input v-model.trim="plugin.name" minlength="1" maxlength="30" show-word-limit
            placeholder="请输入插件名称"></el-input>
        </el-form-item>
        <el-form-item label="插件标题" prop="title">
          <el-input v-model.trim="plugin.title" minlength="1" maxlength="60" show-word-limit
            placeholder="请输入插件标题"></el-input>
        </el-form-item>
        <el-form-item label="插件描述" prop="description">
          <el-input v-model.trim="plugin.description" type="textarea" minlength="1" maxlength="200"
            show-word-limit placeholder="请输入插件描述"></el-input>
        </el-form-item>
        <el-form-item label="版本号" prop="version">
          <el-input v-model.trim="plugin.version" minlength="1" maxlength="20" show-word-limit
            placeholder="请输入版本号"></el-input>
        </el-form-item>
        <el-form-item label="作者" prop="author">
          <el-input v-model.trim="plugin.author" minlength="1" maxlength="30" show-word-limit
            placeholder="请输入作者"></el-input>
        </el-form-item>
        <el-form-item label="入口文件" prop="entry">
          <el-input v-model.trim="plugin.entry" minlength="1" maxlength="100" show-word-limit
            placeholder="请输入入口文件路径"></el-input>
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model.trim="plugin.icon" maxlength="200" show-word-limit
            placeholder="请输入图标URL"></el-input>
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model.trim="plugin.path" minlength="1" maxlength="100" show-word-limit
            placeholder="请输入插件路径"></el-input>
        </el-form-item>
        <el-form-item label="启用" prop="status">
          <el-radio-group v-model="plugin.status">
            <el-radio :value="'enabled'">是</el-radio>
            <el-radio :value="'disabled'">否</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="data.plugin_dialog_visible = false">取消</el-button>
          <el-button type="success" @click="savePlugin">保存</el-button>
        </span>
      </template>
    </el-dialog>
    <!-- ========================================= -->
  </el-tab-pane>
</template>

<script setup lang="ts">
import { onMounted, reactive } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import axios from "axios";

enum Mode {
  "create" = 0,
  "update" = 1,
}

/**
 * 插件信息
 */

interface Plugin {
  id: number;
  name: string;
  title: string;
  description: string;
  version: string;
  author: string;
  entry: string;
  icon: string;
  path: string;
  status: string;
  created_at?: string;
  updated_at?: string;
}

interface ResponseData {
  code: number;
  msg: string;
  data?: any
}

let data = reactive({
  mode: Mode.create,
  plugin_dialog_visible: false,
  active_name: "pluginManage",
  plugin_list: Array<Plugin>(),
});

let plugin = reactive<Plugin>({
  id: 0,
  name: "",
  title: "",
  description: "",
  version: "1.0.0",
  author: "",
  entry: "index.html",
  icon: "",
  path: "",
  status: "enabled"
});

/**
 * 组件挂载时获取插件列表
 */
onMounted(() => {
  getPluginList(0, 10000);
});

/**
 * 获取插件列表
 */
function getPluginList(offset: number = 0, limit: number = 10000) {
  axios.get<ResponseData>(`/api/plugin?offset=${offset}&limit=${limit}`)
    .then((ret) => {
      if (ret.data.code === 0) {
        // 确保status字段是字符串类型
        const processedPlugins = ret.data.data.map((plugin: any) => ({
          ...plugin,
          status: typeof plugin.status === 'number' 
            ? (plugin.status === 1 ? 'enabled' : 'disabled') 
            : plugin.status
        }));
        data.plugin_list = processedPlugins;
      } else {
        ElMessage.error("获取插件列表错误: " + ret.data.msg);
      }
    })
    .catch((error) => {
      ElMessage.error("获取插件列表失败: " + error.message);
    });
}

/**
 * 添加插件
 */
function addPlugin() {
  plugin.id = 0;
  plugin.name = "";
  plugin.title = "";
  plugin.description = "";
  plugin.version = "1.0.0";
  plugin.author = "";
  plugin.entry = "index.html";
  plugin.icon = "";
  plugin.path = "";
  plugin.status = "enabled";
  data.plugin_dialog_visible = true;
}

/**
 * 编辑插件
 * @param p 
 */
function editPlugin(p: Plugin) {
    plugin.id = p.id;
    plugin.name = p.name;
    plugin.title = p.title;
    plugin.description = p.description;
    plugin.version = p.version;
    plugin.author = p.author;
    plugin.entry = p.entry;
    plugin.icon = p.icon;
    plugin.path = p.path;
    // 确保status是字符串类型
    plugin.status = typeof p.status === 'number' ? (p.status === 1 ? 'enabled' : 'disabled') : p.status;
    data.plugin_dialog_visible = true;
  }

/**
 * 保存插件
 */
function savePlugin() {
  // 验证必填字段
  if (!plugin.name) {
    ElMessage.error("请输入插件名称");
    return;
  }
  if (!plugin.title) {
    ElMessage.error("请输入插件标题");
    return;
  }
  if (!plugin.path) {
    ElMessage.error("请输入插件路径");
    return;
  }

  const requestBody = {
    id: plugin.id,
    name: plugin.name,
    title: plugin.title,
    description: plugin.description,
    version: plugin.version,
    author: plugin.author,
    entry_file: plugin.entry,  // 后端需要entry_file字段
    icon: plugin.icon,
    path: plugin.path,
    status: plugin.status,
    order_num: 1  // 改回order_num，与后端JSON标签一致
  };

  if (plugin.id === 0) {
    // 创建新插件
    axios.post<ResponseData>('/api/plugin', requestBody)
      .then((ret) => {
        if (ret.data.code === 0) {
          ElMessage.success("添加插件成功");
          data.plugin_dialog_visible = false;
          getPluginList(0, 10000);
        } else {
          ElMessage.error("添加插件失败: " + ret.data.msg);
        }
      })
      .catch((error) => {
        ElMessage.error("添加插件失败: " + error.message);
      });
  } else {
    // 更新插件
    axios.put<ResponseData>('/api/plugin', requestBody)
      .then((ret) => {
        if (ret.data.code === 0) {
          ElMessage.success("更新插件成功");
          data.plugin_dialog_visible = false;
          getPluginList(0, 10000);
        } else {
          ElMessage.error("更新插件失败: " + ret.data.msg);
        }
      })
      .catch((error) => {
        ElMessage.error("更新插件失败: " + error.message);
      });
  }
}

/**
 * 根据ID删除插件
 * @param pluginId ID
 */
function deletePluginById(pluginId: number) {
  axios.delete<ResponseData>(`/api/plugin/${pluginId}`)
    .then((ret) => {
      if (ret.data.code === 0) {
        ElMessage.success("删除插件成功");
        getPluginList(0, 10000);
      } else {
        ElMessage.error("删除插件失败: " + ret.data.msg);
      }
    })
    .catch((error) => {
      ElMessage.error("删除插件失败: " + error.message);
    });
}

/**
 * 切换插件状态
 * @param p 插件信息
 */
function togglePluginStatus(p: Plugin) {
    const newStatus = p.status === 'enabled' ? 'disabled' : 'enabled';
    const statusText = p.status === 'enabled' ? "禁用" : "启用";
  
  ElMessageBox.confirm(`确定要${statusText}该插件吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    axios.patch<ResponseData>(`/api/plugin/${p.id}/status`)
      .then((ret) => {
        if (ret.data.code === 0) {
          ElMessage.success(`插件${statusText}成功`);
          getPluginList(0, 10000);
        } else {
          ElMessage.error(`插件${statusText}失败: ` + ret.data.msg);
        }
      })
      .catch((error) => {
        ElMessage.error(`插件${statusText}失败: ` + error.message);
      });
  }).catch(() => {
    // 用户取消操作
  });
}

/**
 * 预览/访问插件
 * @param p 插件信息
 */
function viewPlugin(p: Plugin) {
    if (p.status === 'disabled') {
      ElMessage.warning("请先启用插件");
      return;
    }
  
  // 打开插件页面，带上插件ID参数和认证token
  const token = localStorage.getItem("token");
  const pluginUrl = `/plugin/${p.name}?plugin_id=${p.id}&Authorization=${encodeURIComponent(token || '')}`;
  window.open(pluginUrl, '_blank');
}
</script>

<style scoped>
/* 可以添加自定义样式 */
</style>