import { userService } from '@/services/user';
import { roleService } from '@/services/role';
import { RoleUtils } from '@/utils/roleUtils';
import { User, Role } from '@/types';

/**
 * 用户角色管理功能测试工具
 */
export class RoleManagementTest {
  /**
   * 测试角色ID验证功能
   */
  static testRoleIdValidation(): void {
    console.group('🧪 测试角色ID验证功能');
    
    // 测试空数组
    const emptyTest = RoleUtils.validateRoleIds([]);
    console.log('空数组测试:', emptyTest);
    
    // 测试重复ID
    const duplicateTest = RoleUtils.validateRoleIds([
      '123e4567-e89b-12d3-a456-426614174000',
      '123e4567-e89b-12d3-a456-426614174000'
    ]);
    console.log('重复ID测试:', duplicateTest);
    
    // 测试无效UUID格式
    const invalidUuidTest = RoleUtils.validateRoleIds(['invalid-uuid', 'another-invalid']);
    console.log('无效UUID测试:', invalidUuidTest);
    
    // 测试有效UUID
    const validTest = RoleUtils.validateRoleIds([
      '123e4567-e89b-12d3-a456-426614174000',
      '987fcdeb-51a2-43d1-9f12-345678901234'
    ]);
    console.log('有效UUID测试:', validTest);
    
    console.groupEnd();
  }
  
  /**
   * 测试角色数组比较功能
   */
  static testRoleArrayComparison(): void {
    console.group('🧪 测试角色数组比较功能');
    
    const roles1 = ['a', 'b', 'c'];
    const roles2 = ['c', 'a', 'b']; // 顺序不同但内容相同
    const roles3 = ['a', 'b', 'd']; // 内容不同
    
    console.log('相同内容不同顺序:', RoleUtils.compareRoleArrays(roles1, roles2));
    console.log('不同内容:', RoleUtils.compareRoleArrays(roles1, roles3));
    console.log('相同数组:', RoleUtils.compareRoleArrays(roles1, roles1));
    
    console.groupEnd();
  }
  
  /**
   * 测试角色变更检测功能
   */
  static testRoleChangesDetection(): void {
    console.group('🧪 测试角色变更检测功能');
    
    const oldRoles = ['role1', 'role2', 'role3'];
    const newRoles = ['role2', 'role4', 'role5'];
    
    const changes = RoleUtils.getRoleChanges(oldRoles, newRoles);
    console.log('角色变更检测:', changes);
    
    // 测试格式化变更信息
    const mockRoles: Role[] = [
      { id: 'role1', name: '管理员', description: '系统管理员', isSystem: true, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role2', name: '编辑者', description: '内容编辑者', isSystem: false, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role3', name: '查看者', description: '只读用户', isSystem: false, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role4', name: '审核员', description: '内容审核员', isSystem: false, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role5', name: '操作员', description: '系统操作员', isSystem: true, createdAt: new Date(), updatedAt: new Date() }
    ];
    
    const formattedChanges = RoleUtils.formatRoleChanges(changes, mockRoles);
    console.log('格式化变更信息:', formattedChanges);
    
    console.groupEnd();
  }
  
  /**
   * 测试角色存在性验证
   */
  static testRoleExistenceValidation(): void {
    console.group('🧪 测试角色存在性验证');
    
    const availableRoles: Role[] = [
      { id: 'role1', name: '管理员', description: '系统管理员', isSystem: true, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role2', name: '编辑者', description: '内容编辑者', isSystem: false, createdAt: new Date(), updatedAt: new Date() }
    ];
    
    const selectedRoles1 = ['role1', 'role2']; // 都存在
    const selectedRoles2 = ['role1', 'role3']; // role3不存在
    
    const validation1 = RoleUtils.validateRolesExist(selectedRoles1, availableRoles);
    const validation2 = RoleUtils.validateRolesExist(selectedRoles2, availableRoles);
    
    console.log('全部存在的角色验证:', validation1);
    console.log('部分不存在的角色验证:', validation2);
    
    console.groupEnd();
  }
  
  /**
   * 测试系统角色过滤功能
   */
  static testSystemRoleFiltering(): void {
    console.group('🧪 测试系统角色过滤功能');
    
    const mockRoles: Role[] = [
      { id: 'role1', name: '超级管理员', description: '系统超级管理员', isSystem: true, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role2', name: '管理员', description: '系统管理员', isSystem: true, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role3', name: '编辑者', description: '内容编辑者', isSystem: false, createdAt: new Date(), updatedAt: new Date() },
      { id: 'role4', name: '查看者', description: '只读用户', isSystem: false, createdAt: new Date(), updatedAt: new Date() }
    ];
    
    const systemRoles = RoleUtils.filterSystemRoles(mockRoles);
    const normalRoles = RoleUtils.filterNormalRoles(mockRoles);
    
    console.log('系统角色:', systemRoles.map(r => r.name));
    console.log('普通角色:', normalRoles.map(r => r.name));
    
    const mockUser: User = {
      id: 'user1',
      username: 'testuser',
      email: 'test@example.com',
      roles: [mockRoles[0], mockRoles[2]], // 包含一个系统角色
      status: 'active',
      createdAt: new Date(),
      updatedAt: new Date()
    };
    
    console.log('用户是否有系统角色:', RoleUtils.hasSystemRole(mockUser));
    
    console.groupEnd();
  }
  
  /**
   * 运行所有测试
   */
  static runAllTests(): void {
    console.group('🚀 用户角色管理功能测试');
    console.log('开始测试用户角色管理功能的各项优化...');
    
    this.testRoleIdValidation();
    this.testRoleArrayComparison();
    this.testRoleChangesDetection();
    this.testRoleExistenceValidation();
    this.testSystemRoleFiltering();
    
    console.log('✅ 所有测试完成！');
    console.groupEnd();
  }
  
  /**
   * 模拟角色分配流程测试
   */
  static async simulateRoleAssignmentFlow(): Promise<void> {
    console.group('🎭 模拟角色分配流程测试');
    
    try {
      // 注意：这是模拟测试，实际环境中需要真实的用户和角色数据
      console.log('1. 获取用户列表...');
      // const users = await userService.getUsers({ page: 1, size: 10 });
      
      console.log('2. 获取角色列表...');
      // const roles = await roleService.getRoles({ page: 1, size: 20 });
      
      console.log('3. 模拟角色分配验证...');
      const mockRoleIds = [
        '123e4567-e89b-12d3-a456-426614174000',
        '987fcdeb-51a2-43d1-9f12-345678901234'
      ];
      
      const validation = RoleUtils.validateRoleIds(mockRoleIds);
      console.log('角色ID验证结果:', validation);
      
      if (validation.isValid) {
        console.log('✅ 角色分配验证通过');
        // 在实际环境中，这里会调用 userService.assignRoles
        console.log('4. 执行角色分配...');
        console.log('5. 验证分配结果...');
        console.log('✅ 角色分配流程测试完成');
      } else {
        console.log('❌ 角色分配验证失败:', validation.errors);
      }
      
    } catch (error) {
      console.error('❌ 角色分配流程测试失败:', error);
    }
    
    console.groupEnd();
  }
}

// 导出测试函数供控制台使用
(window as any).RoleManagementTest = RoleManagementTest;

// 自动运行基础测试
if (process.env.NODE_ENV === 'development') {
  console.log('🔧 开发环境检测到，自动运行角色管理测试...');
  RoleManagementTest.runAllTests();
}