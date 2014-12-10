#include "clar_libgit2.h"

#include "git2/clone.h"
#include "clone.h"
#include "buffer.h"
#include "path.h"
#include "posix.h"
#include "fileops.h"

void test_clone_local__should_clone_local(void)
{
	git_buf buf = GIT_BUF_INIT;
	const char *path;

	/* we use a fixture path because it needs to exist for us to want to clone */
	
	cl_git_pass(git_buf_printf(&buf, "file://%s", cl_fixture("testrepo.git")));
	cl_assert_equal_i(false, git_clone__should_clone_local(buf.ptr, GIT_CLONE_LOCAL_AUTO));
	cl_assert_equal_i(true,  git_clone__should_clone_local(buf.ptr, GIT_CLONE_LOCAL));
	cl_assert_equal_i(true,  git_clone__should_clone_local(buf.ptr, GIT_CLONE_LOCAL_NO_LINKS));
	cl_assert_equal_i(false, git_clone__should_clone_local(buf.ptr, GIT_CLONE_NO_LOCAL));
	git_buf_free(&buf);

	path = cl_fixture("testrepo.git");
	cl_assert_equal_i(true,  git_clone__should_clone_local(path, GIT_CLONE_LOCAL_AUTO));
	cl_assert_equal_i(true,  git_clone__should_clone_local(path, GIT_CLONE_LOCAL));
	cl_assert_equal_i(true,  git_clone__should_clone_local(path, GIT_CLONE_LOCAL_NO_LINKS));
	cl_assert_equal_i(false, git_clone__should_clone_local(path, GIT_CLONE_NO_LOCAL));
}

void test_clone_local__hardlinks(void)
{
	git_repository *repo;
	git_remote *remote;
	git_signature *sig;
	git_buf buf = GIT_BUF_INIT;
	struct stat st;


	/*
	 * In this first clone, we just copy over, since the temp dir
	 * will often be in a different filesystem, so we cannot
	 * link. It also allows us to control the number of links
	 */
	cl_git_pass(git_repository_init(&repo, "./clone.git", true));
	cl_git_pass(git_remote_create(&remote, repo, "origin", cl_fixture("testrepo.git")));
	cl_git_pass(git_signature_now(&sig, "foo", "bar"));
	cl_git_pass(git_clone_local_into(repo, remote, NULL, NULL, false, sig));

	git_remote_free(remote);
	git_repository_free(repo);

	/* This second clone is in the same filesystem, so we can hardlink */

	cl_git_pass(git_repository_init(&repo, "./clone2.git", true));
	cl_git_pass(git_buf_puts(&buf, cl_git_path_url("clone.git")));
	cl_git_pass(git_remote_create(&remote, repo, "origin", buf.ptr));
	cl_git_pass(git_clone_local_into(repo, remote, NULL, NULL, true, sig));

#ifndef GIT_WIN32
	git_buf_clear(&buf);
	cl_git_pass(git_buf_join_n(&buf, '/', 4, git_repository_path(repo), "objects", "08", "b041783f40edfe12bb406c9c9a8a040177c125"));

	cl_git_pass(p_stat(buf.ptr, &st));
	cl_assert_equal_i(2, st.st_nlink);
#endif

	git_remote_free(remote);
	git_repository_free(repo);
	git_buf_clear(&buf);

	cl_git_pass(git_repository_init(&repo, "./clone3.git", true));
	cl_git_pass(git_buf_puts(&buf, cl_git_path_url("clone.git")));
	cl_git_pass(git_remote_create(&remote, repo, "origin", buf.ptr));
	cl_git_pass(git_clone_local_into(repo, remote, NULL, NULL, false, sig));

	git_buf_clear(&buf);
	cl_git_pass(git_buf_join_n(&buf, '/', 4, git_repository_path(repo), "objects", "08", "b041783f40edfe12bb406c9c9a8a040177c125"));

	cl_git_pass(p_stat(buf.ptr, &st));
	cl_assert_equal_i(1, st.st_nlink);

	git_remote_free(remote);
	git_repository_free(repo);

	/* this one should automatically use links */
	cl_git_pass(git_clone(&repo, "./clone.git", "./clone4.git", NULL));

#ifndef GIT_WIN32
	git_buf_clear(&buf);
	cl_git_pass(git_buf_join_n(&buf, '/', 4, git_repository_path(repo), "objects", "08", "b041783f40edfe12bb406c9c9a8a040177c125"));

	cl_git_pass(p_stat(buf.ptr, &st));
	cl_assert_equal_i(3, st.st_nlink);
#endif

	git_buf_free(&buf);
	git_signature_free(sig);
	git_repository_free(repo);

	cl_git_pass(git_futils_rmdir_r("./clone.git", NULL, GIT_RMDIR_REMOVE_FILES));
	cl_git_pass(git_futils_rmdir_r("./clone2.git", NULL, GIT_RMDIR_REMOVE_FILES));
	cl_git_pass(git_futils_rmdir_r("./clone3.git", NULL, GIT_RMDIR_REMOVE_FILES));
	cl_git_pass(git_futils_rmdir_r("./clone4.git", NULL, GIT_RMDIR_REMOVE_FILES));
}
